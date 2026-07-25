package gdrive

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

const scopeUserInfoEmail = "https://www.googleapis.com/auth/userinfo.email"

func (s *Service) oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.cfg.ClientID,
		ClientSecret: s.cfg.ClientSecret,
		RedirectURL:  s.cfg.RedirectURL,
		Scopes: []string{
			drive.DriveFileScope,
			scopeUserInfoEmail,
		},
		Endpoint: google.Endpoint,
	}
}

func (s *Service) authCodeURL(state string) string {
	return s.oauthConfig().AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

func (s *Service) exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.oauthConfig().Exchange(ctx, code)
}
