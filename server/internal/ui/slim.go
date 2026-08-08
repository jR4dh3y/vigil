//go:build slim

package ui

import (
	"html/template"
	"net/http"
	"strings"
)

// Handler serves a secure connection page that deep-links to the hosted
// dashboard with a server query parameter. When no hosted dashboard URL is
// configured it renders an actionable configuration error instead of a dead
// link. The slim build embeds no dashboard assets.
func Handler(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		href := BuildDashboardURL(cfg.HostedDashboardURL, cfg.PublicURL, r)
		if href == "" {
			serveMissingHosted(w)
			return
		}
		serveConnectionPage(w, href)
	})
}

const missingTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>NVR &mdash; dashboard not configured</title>
<style>body{font-family:system-ui,sans-serif;max-width:36rem;margin:4rem auto;padding:0 1rem;line-height:1.5}code{background:#f1f1f1;padding:.1rem .35rem;border-radius:.25rem}</style>
</head>
<body>
<h1>Dashboard not configured</h1>
<p>This headless NVR server has no hosted dashboard URL configured.</p>
<p>Set <code>NVR_HOSTED_DASHBOARD_URL</code> or run <code>nvrd setup --hosted-url &lt;url&gt;</code> on this server, then reload this page.</p>
</body>
</html>
`

const connectionTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Open NVR dashboard</title>
<style>body{font-family:system-ui,sans-serif;max-width:36rem;margin:4rem auto;padding:0 1rem;line-height:1.5}a{display:inline-block;margin-top:1rem;padding:.6rem 1.2rem;background:#111;color:#fff;text-decoration:none;border-radius:.375rem}</style>
</head>
<body>
<h1>NVR server ready</h1>
<p>This server is running headless. Open the hosted dashboard to manage it:</p>
<p><a href="{{.Href}}">Open dashboard</a></p>
</body>
</html>
`

func serveMissingHosted(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(missingTemplate))
}

func serveConnectionPage(w http.ResponseWriter, href string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = template.Must(template.New("connection").Parse(connectionTemplate)).Execute(w, struct{ Href string }{Href: href})
}
