package push

import "context"

// Notification holds the data for a push notification.
type Notification struct {
	DeviceToken string
	Title       string
	Body        string
	Data        map[string]string
}

// Service is the interface for sending push notifications.
type Service interface {
	// Send delivers a push notification to the given device token.
	Send(ctx context.Context, n Notification) error
}
