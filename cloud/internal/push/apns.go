package push

import (
	"context"
	"log"
)

// APNsService is a stub APNs push notification implementation.
// Replace with a real APNs HTTP/2 client when credentials are available.
type APNsService struct{}

// NewAPNs returns a stub APNs service.
func NewAPNs() *APNsService { return &APNsService{} }

// Send logs the notification and returns success.
func (a *APNsService) Send(_ context.Context, n Notification) error {
	log.Printf("[apns stub] Send: token=%s title=%q body=%q data=%v",
		n.DeviceToken, n.Title, n.Body, n.Data)
	return nil
}
