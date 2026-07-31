package gdrive

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

const scopeUserInfoEmail = "https://www.googleapis.com/auth/userinfo.email"

func oauthConfig(cfg Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes: []string{
			drive.DriveFileScope,
			scopeUserInfoEmail,
		},
		Endpoint: google.Endpoint,
	}
}

func (s *Service) authCodeURL(cfg Config, state string) string {
	return oauthConfig(cfg).AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

func (s *Service) exchange(ctx context.Context, cfg Config, code string) (*oauth2.Token, error) {
	return oauthConfig(cfg).Exchange(ctx, code)
}
