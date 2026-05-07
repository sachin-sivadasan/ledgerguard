package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type appLookupError struct {
	statusCode int
	message    string
}

// resolvePartnerAccount resolves partner account from org context (preferred)
// or falls back to user-based lookup.
func resolvePartnerAccount(r *http.Request, partnerRepo repository.PartnerAccountRepository) (*entity.PartnerAccount, *appLookupError) {
	org := middleware.OrgFromContext(r.Context())
	if org != nil {
		account, err := partnerRepo.FindByOrgID(r.Context(), org.ID)
		if err != nil {
			return nil, &appLookupError{http.StatusNotFound, "no partner account for organization"}
		}
		return account, nil
	}
	// Fallback: user-based (for routes without OrgContextMW)
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		return nil, &appLookupError{http.StatusUnauthorized, "authentication required"}
	}
	account, err := partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		return nil, &appLookupError{http.StatusNotFound, "no partner account found"}
	}
	return account, nil
}

// lookupAppID extracts the authenticated user, resolves their partner account,
// and maps the URL's numeric appID to the internal UUID.
func lookupAppID(r *http.Request, partnerRepo repository.PartnerAccountRepository, appRepo repository.AppRepository) (uuid.UUID, *appLookupError) {
	partnerAccount, lookupErr := resolvePartnerAccount(r, partnerRepo)
	if lookupErr != nil {
		return uuid.Nil, lookupErr
	}

	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		return uuid.Nil, &appLookupError{statusCode: http.StatusBadRequest, message: "app ID is required"}
	}
	fullAppGID := appGIDPrefix + appIDStr

	app, err := appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, fullAppGID)
	if err != nil {
		return uuid.Nil, &appLookupError{statusCode: http.StatusNotFound, message: "app not found"}
	}

	return app.ID, nil
}
