package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sachin-sivadasan/ledgerguard/internal/application/service"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"github.com/sachin-sivadasan/ledgerguard/internal/interfaces/http/middleware"
)

// DeviceHandler handles device token registration endpoints
type DeviceHandler struct {
	notificationService *service.NotificationService
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(notificationService *service.NotificationService) *DeviceHandler {
	return &DeviceHandler{notificationService: notificationService}
}

// RegisterDeviceRequest represents the request body for device registration
type RegisterDeviceRequest struct {
	DeviceToken string `json:"device_token"`
	Platform    string `json:"platform"` // "ios", "android", "web"
}

// RegisterDevice registers a device token for push notifications
// POST /api/v1/devices
func (h *DeviceHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DeviceToken == "" {
		writeJSONError(w, http.StatusBadRequest, "device_token is required")
		return
	}

	if req.Platform == "" {
		writeJSONError(w, http.StatusBadRequest, "platform is required")
		return
	}

	platform := entity.Platform(req.Platform)
	if !platform.IsValid() {
		writeJSONError(w, http.StatusBadRequest, "Invalid platform. Must be one of: ios, android, web")
		return
	}

	if err := h.notificationService.RegisterDevice(r.Context(), user.ID, req.DeviceToken, platform); err != nil {
		if errors.Is(err, service.ErrInvalidPlatform) {
			writeJSONError(w, http.StatusBadRequest, "Invalid platform")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to register device")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Device registered successfully",
	})
}

// UnregisterDeviceRequest represents the request body for device unregistration
type UnregisterDeviceRequest struct {
	DeviceToken string `json:"device_token"`
}

// UnregisterDevice removes a device token
// DELETE /api/v1/devices
func (h *DeviceHandler) UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UnregisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DeviceToken == "" {
		writeJSONError(w, http.StatusBadRequest, "device_token is required")
		return
	}

	if err := h.notificationService.UnregisterDevice(r.Context(), user.ID, req.DeviceToken); err != nil {
		if errors.Is(err, service.ErrDeviceTokenNotFound) {
			writeJSONError(w, http.StatusNotFound, "Device token not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to unregister device")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
