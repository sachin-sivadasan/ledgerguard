package external

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
	"google.golang.org/api/option"
)

type FirebaseAuthService struct {
	client *auth.Client
	// checkRevoked additionally rejects revoked/disabled sessions (signed-out users,
	// disabled accounts) instead of trusting a token until its ~1h natural expiry.
	// Costs one Firebase GetUser call per verification.
	checkRevoked bool
}

func NewFirebaseAuthService(ctx context.Context, credentialsFile string, checkRevoked bool) (*FirebaseAuthService, error) {
	var app *firebase.App
	var err error

	if credentialsFile != "" {
		opt := option.WithCredentialsFile(credentialsFile)
		app, err = firebase.NewApp(ctx, nil, opt)
	} else {
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get firebase auth client: %w", err)
	}

	return &FirebaseAuthService{client: client, checkRevoked: checkRevoked}, nil
}

// RevokeRefreshTokens revokes all refresh tokens for the user, so their sessions must
// re-authenticate. With CheckRevoked verification enabled, existing ID tokens are
// rejected within their (~1h) TTL.
func (s *FirebaseAuthService) RevokeRefreshTokens(ctx context.Context, uid string) error {
	if err := s.client.RevokeRefreshTokens(ctx, uid); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}
	return nil
}

func (s *FirebaseAuthService) VerifyIDToken(ctx context.Context, idToken string) (*service.TokenClaims, error) {
	var token *auth.Token
	var err error
	if s.checkRevoked {
		token, err = s.client.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	} else {
		token, err = s.client.VerifyIDToken(ctx, idToken)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	email, _ := token.Claims["email"].(string)
	emailVerified, _ := token.Claims["email_verified"].(bool)

	return &service.TokenClaims{
		UID:           token.UID,
		Email:         email,
		EmailVerified: emailVerified,
	}, nil
}
