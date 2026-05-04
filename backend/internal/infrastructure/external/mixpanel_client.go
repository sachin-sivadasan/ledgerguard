package external

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/sachin-sivadasan/ledgerguard/internal/domain/service"
)

const (
	mixpanelTrackURL  = "https://api.mixpanel.com/track"
	mixpanelEngageURL = "https://api.mixpanel.com/engage"
)

// MixpanelClient sends events to Mixpanel via the HTTP API.
// All calls are fire-and-forget (async goroutine).
type MixpanelClient struct {
	token      string
	httpClient *http.Client
}

// NewMixpanelClient creates a new Mixpanel event tracker.
func NewMixpanelClient(token string) *MixpanelClient {
	return &MixpanelClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Track sends an event to Mixpanel asynchronously.
func (c *MixpanelClient) Track(_ context.Context, userID string, event string, properties service.EventProperties) {
	props := make(map[string]interface{})
	for k, v := range properties {
		props[k] = v
	}
	props["token"] = c.token
	props["distinct_id"] = userID
	props["time"] = time.Now().Unix()

	payload := []map[string]interface{}{
		{
			"event":      event,
			"properties": props,
		},
	}

	go c.send(mixpanelTrackURL, payload)
}

// SetUserProperties sets user profile properties in Mixpanel asynchronously.
func (c *MixpanelClient) SetUserProperties(_ context.Context, userID string, properties service.EventProperties) {
	payload := []map[string]interface{}{
		{
			"$token":       c.token,
			"$distinct_id": userID,
			"$set":         properties,
		},
	}

	go c.send(mixpanelEngageURL, payload)
}

func (c *MixpanelClient) send(url string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("mixpanel: failed to marshal payload: %v", err)
		return
	}

	// Mixpanel expects base64-encoded JSON in the data parameter for GET,
	// or raw JSON array for POST with Content-Type application/json
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		log.Printf("mixpanel: failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("mixpanel: request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("mixpanel: unexpected status %d for %s", resp.StatusCode, url)
	}
}

// Ensure compile-time interface compliance
var _ service.EventTracker = (*MixpanelClient)(nil)
