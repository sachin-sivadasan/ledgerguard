package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// Encryptor encrypts and decrypts sensitive values such as access tokens.
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    http.StatusText(status),
			"message": message,
		},
	})
}

// contextWithUser is a helper for testing
func contextWithUser(ctx context.Context, user *entity.User) context.Context {
	return middleware.SetUserContext(ctx, user)
}

// contextWithOrg is a helper for testing
func contextWithOrg(ctx context.Context, org *entity.Organization) context.Context {
	return middleware.SetOrgContext(ctx, org)
}
