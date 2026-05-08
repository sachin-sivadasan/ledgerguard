package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

const appGIDPrefix = "gid://partners/App/"

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

// resolveAppFromRequest extracts the app ID from the URL parameter "appID",
// accepts either UUID or numeric Shopify ID, and returns the resolved app entity.
// UUID uses fast PK lookup; numeric falls back to GID construction + partner-scoped lookup.
func resolveAppFromRequest(r *http.Request, partnerRepo repository.PartnerAccountRepository, appRepo repository.AppRepository) (*entity.App, *appLookupError) {
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		return nil, &appLookupError{http.StatusBadRequest, "app ID is required"}
	}

	// Fast path: UUID (primary key lookup)
	if appID, err := uuid.Parse(appIDStr); err == nil {
		app, err := appRepo.FindByID(r.Context(), appID)
		if err == nil && app != nil {
			return app, nil
		}
	}

	// Fallback: numeric → GID → FindByPartnerAppID
	partnerAccount, lookupErr := resolvePartnerAccount(r, partnerRepo)
	if lookupErr != nil {
		return nil, lookupErr
	}
	fullGID := appGIDPrefix + appIDStr
	app, err := appRepo.FindByPartnerAppID(r.Context(), partnerAccount.ID, fullGID)
	if err != nil || app == nil {
		return nil, &appLookupError{http.StatusNotFound, "app not found"}
	}
	return app, nil
}
