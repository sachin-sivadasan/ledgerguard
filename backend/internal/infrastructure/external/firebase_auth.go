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
}

func NewFirebaseAuthService(ctx context.Context, credentialsFile string) (*FirebaseAuthService, error) {
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

	return &FirebaseAuthService{client: client}, nil
}

func (s *FirebaseAuthService) VerifyIDToken(ctx context.Context, idToken string) (*service.TokenClaims, error) {
	token, err := s.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	email, _ := token.Claims["email"].(string)

	return &service.TokenClaims{
		UID:   token.UID,
		Email: email,
	}, nil
}
