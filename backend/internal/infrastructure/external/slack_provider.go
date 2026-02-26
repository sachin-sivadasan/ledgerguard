package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrSlackWebhookFailed is returned when the Slack webhook request fails
var ErrSlackWebhookFailed = errors.New("slack webhook request failed")

// ErrInvalidWebhookURL is returned when the webhook URL is empty or invalid
var ErrInvalidWebhookURL = errors.New("invalid webhook URL")

// SlackNotifier defines the interface for sending Slack notifications
type SlackNotifier interface {
	// SendSlack sends a message to a Slack webhook
	SendSlack(ctx context.Context, webhookURL string, title string, body string, color string) error
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color  string `json:"color,omitempty"`
	Title  string `json:"title,omitempty"`
	Text   string `json:"text,omitempty"`
	Footer string `json:"footer,omitempty"`
	Ts     int64  `json:"ts,omitempty"`
}

// SlackPayload represents the Slack webhook payload
type SlackPayload struct {
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackNotificationProvider implements SlackNotifier for sending Slack webhooks
type SlackNotificationProvider struct {
	httpClient *http.Client
}

// NewSlackNotificationProvider creates a new Slack notification provider
func NewSlackNotificationProvider() *SlackNotificationProvider {
	return &SlackNotificationProvider{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewSlackNotificationProviderWithClient creates a provider with a custom HTTP client (for testing)
func NewSlackNotificationProviderWithClient(client *http.Client) *SlackNotificationProvider {
	return &SlackNotificationProvider{
		httpClient: client,
	}
}

// SendSlack sends a message to a Slack webhook
func (p *SlackNotificationProvider) SendSlack(ctx context.Context, webhookURL string, title string, body string, color string) error {
	if webhookURL == "" {
		return ErrInvalidWebhookURL
	}

	// Build payload with attachment for rich formatting
	payload := SlackPayload{
		Attachments: []SlackAttachment{
			{
				Color:  color,
				Title:  title,
				Text:   body,
				Footer: "LedgerGuard",
				Ts:     time.Now().Unix(),
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status code %d", ErrSlackWebhookFailed, resp.StatusCode)
	}

	return nil
}

// Slack color constants
const (
	SlackColorDanger  = "#dc3545" // Red - for critical alerts
	SlackColorWarning = "#ffc107" // Yellow - for warnings
	SlackColorSuccess = "#28a745" // Green - for success
	SlackColorInfo    = "#17a2b8" // Blue - for info
)
