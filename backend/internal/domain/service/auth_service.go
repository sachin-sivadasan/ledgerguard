package service

import "context"

type TokenClaims struct {
	UID           string
	Email         string
	EmailVerified bool
}

type AuthTokenVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*TokenClaims, error)
}

// TokenRevoker revokes all of a user's refresh tokens (server-side force-logout).
// Combined with revocation-checked verification, this invalidates existing sessions
// within the ID-token TTL. Satisfied by the Firebase auth service.
type TokenRevoker interface {
	RevokeRefreshTokens(ctx context.Context, uid string) error
}
