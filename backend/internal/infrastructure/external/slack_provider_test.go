package external

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSlackNotificationProvider_SendSlack(t *testing.T) {
	ctx := context.Background()

	t.Run("sends slack message successfully", func(t *testing.T) {
		var receivedPayload SlackPayload

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
				t.Errorf("failed to decode payload: %v", err)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		provider := NewSlackNotificationProviderWithClient(server.Client())
		err := provider.SendSlack(ctx, server.URL, "Test Title", "Test Body", SlackColorDanger)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify payload
		if len(receivedPayload.Attachments) != 1 {
			t.Fatalf("expected 1 attachment, got %d", len(receivedPayload.Attachments))
		}

		attachment := receivedPayload.Attachments[0]
		if attachment.Title != "Test Title" {
			t.Errorf("expected title 'Test Title', got '%s'", attachment.Title)
		}
		if attachment.Text != "Test Body" {
			t.Errorf("expected text 'Test Body', got '%s'", attachment.Text)
		}
		if attachment.Color != SlackColorDanger {
			t.Errorf("expected color '%s', got '%s'", SlackColorDanger, attachment.Color)
		}
		if attachment.Footer != "LedgerGuard" {
			t.Errorf("expected footer 'LedgerGuard', got '%s'", attachment.Footer)
		}
	})

	t.Run("returns error for empty webhook URL", func(t *testing.T) {
		provider := NewSlackNotificationProvider()
		err := provider.SendSlack(ctx, "", "Title", "Body", SlackColorInfo)

		if !errors.Is(err, ErrInvalidWebhookURL) {
			t.Errorf("expected ErrInvalidWebhookURL, got %v", err)
		}
	})

	t.Run("returns error for non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		provider := NewSlackNotificationProviderWithClient(server.Client())
		err := provider.SendSlack(ctx, server.URL, "Title", "Body", SlackColorWarning)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrSlackWebhookFailed) {
			t.Errorf("expected ErrSlackWebhookFailed, got %v", err)
		}
	})

	t.Run("returns error for invalid URL", func(t *testing.T) {
		provider := NewSlackNotificationProvider()
		err := provider.SendSlack(ctx, "not-a-valid-url", "Title", "Body", SlackColorInfo)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("handles server timeout gracefully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't respond - let it hang
			select {}
		}))
		defer server.Close()

		// Use a client with very short timeout
		shortTimeoutClient := &http.Client{
			Timeout: 1, // 1 nanosecond
		}

		provider := NewSlackNotificationProviderWithClient(shortTimeoutClient)
		err := provider.SendSlack(ctx, server.URL, "Title", "Body", SlackColorInfo)

		if err == nil {
			t.Fatal("expected timeout error, got nil")
		}
	})
}

func TestSlackColors(t *testing.T) {
	t.Run("color constants are valid hex codes", func(t *testing.T) {
		colors := []string{SlackColorDanger, SlackColorWarning, SlackColorSuccess, SlackColorInfo}

		for _, color := range colors {
			if len(color) != 7 || color[0] != '#' {
				t.Errorf("invalid color format: %s", color)
			}
		}
	})
}
