package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/infrastructure/persistence"
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
			if isNotFoundError(err) {
				return nil, &appLookupError{http.StatusNotFound, "no partner account for organization"}
			}
			log.Printf("app_lookup: DB error resolving partner account by org %s: %v", org.ID, err)
			return nil, &appLookupError{http.StatusServiceUnavailable, "service temporarily unavailable"}
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
		if isNotFoundError(err) {
			return nil, &appLookupError{http.StatusNotFound, "no partner account found"}
		}
		log.Printf("app_lookup: DB error resolving partner account by user %s: %v", user.ID, err)
		return nil, &appLookupError{http.StatusServiceUnavailable, "service temporarily unavailable"}
	}
	return account, nil
}

// resolveAppFromRequest extracts the UUID app ID from the URL parameter "appID"
// and returns the resolved app entity via primary key lookup.
func resolveAppFromRequest(r *http.Request, partnerRepo repository.PartnerAccountRepository, appRepo repository.AppRepository) (*entity.App, *appLookupError) {
	appIDStr := chi.URLParam(r, "appID")
	if appIDStr == "" {
		return nil, &appLookupError{http.StatusBadRequest, "app ID is required"}
	}

	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		return nil, &appLookupError{http.StatusBadRequest, "invalid app ID format"}
	}

	app, err := appRepo.FindByID(r.Context(), appID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, &appLookupError{http.StatusNotFound, "app not found"}
		}
		log.Printf("app_lookup: DB error resolving app %s: %v", appID, err)
		return nil, &appLookupError{http.StatusServiceUnavailable, "service temporarily unavailable"}
	}
	if app == nil {
		return nil, &appLookupError{http.StatusNotFound, "app not found"}
	}
	return app, nil
}

// isNotFoundError checks if the error is a known "not found" sentinel error.
func isNotFoundError(err error) bool {
	return errors.Is(err, persistence.ErrPartnerAccountNotFound) ||
		errors.Is(err, persistence.ErrAppNotFound)
}
