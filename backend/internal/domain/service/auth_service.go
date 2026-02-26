package service

import "context"

type TokenClaims struct {
	UID   string
	Email string
}

type AuthTokenVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*TokenClaims, error)
}
