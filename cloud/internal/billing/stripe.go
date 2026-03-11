package billing

import (
	"context"
	"fmt"
	"log"
)

// CheckoutSession holds the result of creating a Stripe checkout session.
type CheckoutSession struct {
	SessionID  string
	CheckoutURL string
}

// StripeService is the interface for Stripe billing operations.
type StripeService interface {
	// CreateCheckoutSession creates a Stripe Checkout session for the given user
	// and tier, returning a redirect URL.
	CreateCheckoutSession(ctx context.Context, userID string, tier string) (*CheckoutSession, error)

	// CancelSubscription cancels an active Stripe subscription for a user.
	CancelSubscription(ctx context.Context, userID string) error
}

// StubStripe is a no-op StripeService that logs calls and returns success.
// Replace with a real implementation when Stripe API keys are available.
type StubStripe struct{}

// NewStubStripe returns a StubStripe instance.
func NewStubStripe() *StubStripe { return &StubStripe{} }

func (s *StubStripe) CreateCheckoutSession(_ context.Context, userID, tier string) (*CheckoutSession, error) {
	log.Printf("[stripe stub] CreateCheckoutSession: userID=%s tier=%s", userID, tier)
	return &CheckoutSession{
		SessionID:   fmt.Sprintf("stub_cs_%s_%s", userID, tier),
		CheckoutURL: fmt.Sprintf("https://checkout.stripe.com/stub?user=%s&tier=%s", userID, tier),
	}, nil
}

func (s *StubStripe) CancelSubscription(_ context.Context, userID string) error {
	log.Printf("[stripe stub] CancelSubscription: userID=%s", userID)
	return nil
}
