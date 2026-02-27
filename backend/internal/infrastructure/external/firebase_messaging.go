package external

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/sachin-sivadasan/ledgerguard/internal/domain/entity"
	"google.golang.org/api/option"
)

// FirebaseMessagingService handles sending push notifications via FCM
type FirebaseMessagingService struct {
	client *messaging.Client
}

// NewFirebaseMessagingService creates a new FCM service
func NewFirebaseMessagingService(ctx context.Context, credentialsFile string) (*FirebaseMessagingService, error) {
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging client: %w", err)
	}

	return &FirebaseMessagingService{client: client}, nil
}

// SendPush sends a push notification to a device
func (s *FirebaseMessagingService) SendPush(ctx context.Context, deviceToken string, platform entity.Platform, title string, body string) error {
	message := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	// Add platform-specific configuration
	switch platform {
	case entity.PlatformIOS:
		message.APNS = &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					Badge: intPtr(1),
				},
			},
		}
	case entity.PlatformAndroid:
		message.Android = &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		}
	}

	_, err := s.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}

	return nil
}

// SendMulticast sends a notification to multiple devices
func (s *FirebaseMessagingService) SendMulticast(ctx context.Context, tokens []string, title string, body string) (*messaging.BatchResponse, error) {
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	return s.client.SendEachForMulticast(ctx, message)
}

func intPtr(i int) *int {
	return &i
}
