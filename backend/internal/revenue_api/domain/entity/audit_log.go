package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an API request audit log entry
type AuditLog struct {
	ID             uuid.UUID
	APIKeyID       uuid.UUID
	Endpoint       string
	Method         string
	RequestParams  map[string]interface{}
	ResponseStatus int
	ResponseTimeMs int
	IPAddress      string
	UserAgent      string
	CreatedAt      time.Time
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(
	apiKeyID uuid.UUID,
	endpoint string,
	method string,
	requestParams map[string]interface{},
	responseStatus int,
	responseTimeMs int,
	ipAddress string,
	userAgent string,
) *AuditLog {
	return &AuditLog{
		ID:             uuid.New(),
		APIKeyID:       apiKeyID,
		Endpoint:       endpoint,
		Method:         method,
		RequestParams:  requestParams,
		ResponseStatus: responseStatus,
		ResponseTimeMs: responseTimeMs,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		CreatedAt:      time.Now().UTC(),
	}
}

// RequestParamsJSON returns the request params as JSON bytes
func (a *AuditLog) RequestParamsJSON() ([]byte, error) {
	if a.RequestParams == nil {
		return nil, nil
	}
	return json.Marshal(a.RequestParams)
}
