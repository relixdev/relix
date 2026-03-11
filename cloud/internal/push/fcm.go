package push

import (
	"context"
	"log"
)

// FCMService is a stub Firebase Cloud Messaging push notification implementation.
// Replace with a real FCM HTTP v1 client when credentials are available.
type FCMService struct{}

// NewFCM returns a stub FCM service.
func NewFCM() *FCMService { return &FCMService{} }

// Send logs the notification and returns success.
func (f *FCMService) Send(_ context.Context, n Notification) error {
	log.Printf("[fcm stub] Send: token=%s title=%q body=%q data=%v",
		n.DeviceToken, n.Title, n.Body, n.Data)
	return nil
}
