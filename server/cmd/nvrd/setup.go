package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/nvr/nvr/server/internal/auth"
	"github.com/nvr/nvr/server/internal/bootstrap"
	"github.com/nvr/nvr/server/internal/config"
	"github.com/nvr/nvr/server/internal/store"
	"golang.org/x/term"
)

// errSetupComplete marks an explicit `nvrd setup` against a server that already
// has an admin user.
var errSetupComplete = errors.New("setup already completed: this server already has an admin user")

const setupUsageText = `Usage: nvrd setup [flags]

Configure this NVR server for first use: creates the first admin account and
persists the public and hosted dashboard URLs. URLs are stored in the database;
the admin password is never persisted or shown on the command line.

Flags:
  --username <name>       Admin username (default: $NVR_ADMIN_USERNAME)
  --password-stdin        Read the admin password from standard input instead
                          of prompting on the terminal
  --public-url <url>      Public URL of this server (default: $NVR_PUBLIC_URL)
  --hosted-url <url>      Hosted dashboard URL used to reach this server
                          (default: $NVR_HOSTED_DASHBOARD_URL)
  --non-interactive       Fail instead of prompting when required values are
                          missing
  -h, --help              Show this help
`

// setupOptions carries the inputs and outputs for setup so the logic is
// testable without a real terminal or process environment.
type setupOptions struct {
	// Flag values (empty when not provided).
	Username  string
	PublicURL string
	HostedURL string
	// PasswordStdin reads the password from Stdin instead of the terminal.
	PasswordStdin  bool
	NonInteractive bool
	// Env fallbacks.
	EnvUsername  string
	EnvPassword  string
	EnvPublicURL string
	EnvHostedURL string
	// I/O.
	Stdin  io.Reader
	Stdout io.Writer
	// ReadPassword prompts on a hidden terminal for an interactive password.
	ReadPassword func() (string, error)
}

func runSetup(args []string) int {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, setupUsageText) }
	var (
		username       = fs.String("username", "", "admin username (default: $NVR_ADMIN_USERNAME)")
		passwordStdin  = fs.Bool("password-stdin", false, "read the admin password from standard input")
		publicURL      = fs.String("public-url", "", "public URL of this server (default: $NVR_PUBLIC_URL)")
		hostedURL      = fs.String("hosted-url", "", "hosted dashboard URL (default: $NVR_HOSTED_DASHBOARD_URL)")
		nonInteractive = fs.Bool("non-interactive", false, "fail instead of prompting when values are missing")
	)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		return 1
	}
	if err := cfg.EnsureDirs(); err != nil {
		slog.Error("ensure dirs", "err", err)
		return 1
	}

	db, err := store.Open(cfg.DataDir)
	if err != nil {
		slog.Error("open database", "err", err)
		return 1
	}
	defer db.Close()
	if err := store.Migrate(db); err != nil {
		slog.Error("migrate", "err", err)
		return 1
	}
	q := store.New(db)

	o := setupOptions{
		Username:       *username,
		PublicURL:      *publicURL,
		HostedURL:      *hostedURL,
		PasswordStdin:  *passwordStdin,
		NonInteractive: *nonInteractive,
		EnvUsername:    os.Getenv("NVR_ADMIN_USERNAME"),
		EnvPassword:    os.Getenv("NVR_ADMIN_PASSWORD"),
		EnvPublicURL:   os.Getenv("NVR_PUBLIC_URL"),
		EnvHostedURL:   os.Getenv("NVR_HOSTED_DASHBOARD_URL"),
		Stdin:          os.Stdin,
		Stdout:         os.Stdout,
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		o.ReadPassword = func() (string, error) {
			fmt.Fprint(os.Stderr, "Password (hidden): ")
			b, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(os.Stderr)
			return string(b), err
		}
	}

	if err := runSetupWithDB(context.Background(), q, o); err != nil {
		if errors.Is(err, errSetupComplete) {
			fmt.Fprintln(os.Stderr, err.Error())
			return 1
		}
		slog.Error("setup", "err", err)
		return 1
	}
	return 0
}

// runSetupWithDB performs the setup against an open, migrated store.
func runSetupWithDB(ctx context.Context, q *store.Queries, o setupOptions) error {
	username := firstNonEmpty(o.Username, o.EnvUsername)
	publicURL := firstNonEmpty(o.PublicURL, o.EnvPublicURL)
	hostedURL := firstNonEmpty(o.HostedURL, o.EnvHostedURL)

	for name, val := range map[string]string{"public URL": publicURL, "hosted dashboard URL": hostedURL} {
		if val == "" {
			continue
		}
		if err := bootstrap.ValidateURL(val); err != nil {
			return fmt.Errorf("invalid %s %q: %w", name, val, err)
		}
	}

	// Reject an already-complete setup cleanly when credential-bearing input is
	// supplied; allow URL-only updates (public/hosted dashboard URLs) when no
	// username/password bootstrap input is given.
	count, err := q.CountUsers(ctx)
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}
	if count > 0 {
		if credentialInput(o) {
			return errSetupComplete
		}
		return updateURLSettings(ctx, q, publicURL, hostedURL, o.Stdout)
	}

	password, err := o.resolvePassword()
	if err != nil {
		return err
	}

	if _, err := auth.CreateFirstAdmin(ctx, q, username, password); err != nil {
		return err
	}

	if err := updateURLSettings(ctx, q, publicURL, hostedURL, o.Stdout); err != nil {
		return err
	}

	fmt.Fprintln(o.Stdout, "Setup complete.")
	fmt.Fprintf(o.Stdout, "  Admin username: %s\n", username)
	return nil
}

// credentialInput reports whether the operator supplied first-admin credentials
// (a username or a password source), as opposed to a URL-only update.
func credentialInput(o setupOptions) bool {
	return firstNonEmpty(o.Username, o.EnvUsername) != "" ||
		o.EnvPassword != "" ||
		o.PasswordStdin
}

// updateURLSettings persists the provided non-secret URLs (when non-empty) and
// prints the resulting configuration. It never creates users.
func updateURLSettings(ctx context.Context, q *store.Queries, publicURL, hostedURL string, stdout io.Writer) error {
	if publicURL != "" {
		if err := q.UpsertSetting(ctx, store.UpsertSettingParams{Key: bootstrap.SettingPublicURL, Value: publicURL}); err != nil {
			return fmt.Errorf("persist public URL: %w", err)
		}
	}
	if hostedURL != "" {
		if err := q.UpsertSetting(ctx, store.UpsertSettingParams{Key: bootstrap.SettingHostedDashboardURL, Value: hostedURL}); err != nil {
			return fmt.Errorf("persist hosted dashboard URL: %w", err)
		}
	}
	fmt.Fprintln(stdout, "Settings updated.")
	if publicURL != "" {
		fmt.Fprintf(stdout, "  Public URL: %s\n", publicURL)
	} else {
		fmt.Fprintln(stdout, "  Public URL: (not set)")
	}
	if hostedURL != "" {
		fmt.Fprintf(stdout, "  Hosted dashboard URL: %s\n", hostedURL)
	} else {
		fmt.Fprintln(stdout, "  Hosted dashboard URL: (not set)")
	}
	return nil
}

// resolvePassword picks the password source: an explicit --password-stdin flag
// wins, then NVR_ADMIN_PASSWORD env, then a hidden terminal prompt. In
// non-interactive mode a missing password is an error. The password is never
// read from a CLI flag or a file.
func (o setupOptions) resolvePassword() (string, error) {
	if o.PasswordStdin {
		return readStdinPassword(o.Stdin)
	}
	if o.EnvPassword != "" {
		return o.EnvPassword, nil
	}
	if o.NonInteractive {
		return "", errors.New("password required: set NVR_ADMIN_PASSWORD or pass --password-stdin in non-interactive mode")
	}
	if o.ReadPassword != nil {
		return o.ReadPassword()
	}
	return readStdinPassword(o.Stdin)
}

func readStdinPassword(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read password from standard input: %w", err)
	}
	return strings.TrimRight(string(b), "\r\n"), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
