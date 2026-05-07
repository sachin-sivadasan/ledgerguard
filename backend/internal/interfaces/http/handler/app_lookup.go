package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/repository"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

type appLookupError struct {
	statusCode int
	message    string
}

// lookupAppID extracts the authenticated user, resolves their partner account,
// and maps the URL's numeric appID to the internal UUID.
func lookupAppID(r *http.Request, partnerRepo repository.PartnerAccountRepository, appRepo repository.AppRepository) (uuid.UUID, *appLookupError) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		return uuid.Nil, &appLookupError{statusCode: http.StatusUnauthorized, message: "authentication required"}
	}

	partnerAccount, err := partnerRepo.FindByUserID(r.Context(), user.ID)
	if err != nil {
		return uuid.Nil, &appLookupError{statusCode: http.StatusNotFound, message: "no partner account found"}
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
