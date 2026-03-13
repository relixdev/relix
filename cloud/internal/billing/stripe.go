package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/relixdev/relix/cloud/internal/user"

	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
)

// CheckoutSession holds the result of creating a Stripe checkout session.
type CheckoutSession struct {
	SessionID   string
	CheckoutURL string
}

// PortalSession holds the result of creating a Stripe billing portal session.
type PortalSession struct {
	URL string
}

// SubscriptionDetails holds the current subscription state for a user.
type SubscriptionDetails struct {
	ID                 string    `json:"id"`
	Status             string    `json:"status"`
	Tier               string    `json:"tier"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	CancelAtPeriodEnd  bool      `json:"cancel_at_period_end"`
}

// StripeService is the interface for Stripe billing operations.
type StripeService interface {
	// CreateCheckoutSession creates a Stripe Checkout session for the given user
	// and tier, returning a redirect URL.
	CreateCheckoutSession(ctx context.Context, userID string, tier string) (*CheckoutSession, error)

	// CancelSubscription cancels an active Stripe subscription for a user.
	CancelSubscription(ctx context.Context, userID string) error

	// CreatePortalSession creates a Stripe Customer Portal session for subscription management.
	CreatePortalSession(ctx context.Context, userID string) (*PortalSession, error)

	// GetSubscriptionDetails returns the current Stripe subscription for a user.
	GetSubscriptionDetails(ctx context.Context, userID string) (*SubscriptionDetails, error)

	// HandleWebhook processes a Stripe webhook event, returning nil on success.
	HandleWebhook(ctx context.Context, payload []byte, signature string) error
}

// --- Price ID configuration ---

// PriceConfig maps tiers to Stripe price IDs from environment variables.
type PriceConfig struct {
	PlusMonthly  string
	PlusAnnual   string
	ProMonthly   string
	ProAnnual    string
	TeamMonthly  string
	TeamAnnual   string
}

// LoadPriceConfig reads price IDs from environment variables.
func LoadPriceConfig() PriceConfig {
	return PriceConfig{
		PlusMonthly:  os.Getenv("STRIPE_PRICE_PLUS_MONTHLY"),
		PlusAnnual:   os.Getenv("STRIPE_PRICE_PLUS_ANNUAL"),
		ProMonthly:   os.Getenv("STRIPE_PRICE_PRO_MONTHLY"),
		ProAnnual:    os.Getenv("STRIPE_PRICE_PRO_ANNUAL"),
		TeamMonthly:  os.Getenv("STRIPE_PRICE_TEAM_MONTHLY"),
		TeamAnnual:   os.Getenv("STRIPE_PRICE_TEAM_ANNUAL"),
	}
}

// priceForTier returns the default (monthly) price ID for a given tier.
func (pc PriceConfig) priceForTier(tier string) (string, error) {
	switch user.Tier(tier) {
	case user.TierPlus:
		if pc.PlusMonthly == "" {
			return "", fmt.Errorf("STRIPE_PRICE_PLUS_MONTHLY not configured")
		}
		return pc.PlusMonthly, nil
	case user.TierPro:
		if pc.ProMonthly == "" {
			return "", fmt.Errorf("STRIPE_PRICE_PRO_MONTHLY not configured")
		}
		return pc.ProMonthly, nil
	case user.TierTeam:
		if pc.TeamMonthly == "" {
			return "", fmt.Errorf("STRIPE_PRICE_TEAM_MONTHLY not configured")
		}
		return pc.TeamMonthly, nil
	default:
		return "", fmt.Errorf("no price for tier %q", tier)
	}
}

// --- RealStripe implementation ---

// RealStripe implements StripeService using the Stripe API via stripe-go.
type RealStripe struct {
	userStore      user.Store
	webhookSecret  string
	prices         PriceConfig
	successURL     string
	cancelURL      string
	portalReturnURL string
}

// RealStripeConfig holds the configuration for creating a RealStripe instance.
type RealStripeConfig struct {
	SecretKey      string
	WebhookSecret  string
	UserStore      user.Store
	Prices         PriceConfig
	SuccessURL     string // URL to redirect after successful checkout
	CancelURL      string // URL to redirect after cancelled checkout
	PortalReturnURL string // URL to redirect after portal session
}

// NewRealStripe creates a RealStripe that talks to the Stripe API.
// It sets the global stripe.Key — call this once at startup.
func NewRealStripe(cfg RealStripeConfig) *RealStripe {
	stripe.Key = cfg.SecretKey

	successURL := cfg.SuccessURL
	if successURL == "" {
		successURL = "https://relix.sh/billing/success?session_id={CHECKOUT_SESSION_ID}"
	}
	cancelURL := cfg.CancelURL
	if cancelURL == "" {
		cancelURL = "https://relix.sh/billing/cancel"
	}
	portalReturnURL := cfg.PortalReturnURL
	if portalReturnURL == "" {
		portalReturnURL = "https://relix.sh/billing"
	}

	return &RealStripe{
		userStore:       cfg.UserStore,
		webhookSecret:   cfg.WebhookSecret,
		prices:          cfg.Prices,
		successURL:      successURL,
		cancelURL:       cancelURL,
		portalReturnURL: portalReturnURL,
	}
}

// ensureCustomer retrieves or creates a Stripe customer for the given user.
func (s *RealStripe) ensureCustomer(ctx context.Context, userID string) (string, error) {
	u, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	// Already has a Stripe customer ID.
	if u.StripeCustomerID != "" {
		return u.StripeCustomerID, nil
	}

	// Create a new Stripe customer.
	params := &stripe.CustomerParams{
		Email: stripe.String(u.Email),
	}
	params.AddMetadata("relix_user_id", userID)
	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe create customer: %w", err)
	}

	// Persist the customer ID on the user record.
	u.StripeCustomerID = cust.ID
	if _, err := s.userStore.UpdateUser(ctx, u); err != nil {
		return "", fmt.Errorf("save stripe customer id: %w", err)
	}

	return cust.ID, nil
}

func (s *RealStripe) CreateCheckoutSession(ctx context.Context, userID, tier string) (*CheckoutSession, error) {
	priceID, err := s.prices.priceForTier(tier)
	if err != nil {
		return nil, err
	}

	customerID, err := s.ensureCustomer(ctx, userID)
	if err != nil {
		return nil, err
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.successURL),
		CancelURL:  stripe.String(s.cancelURL),
	}
	params.AddMetadata("relix_user_id", userID)
	params.AddMetadata("relix_tier", tier)

	sess, err := checkoutsession.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create checkout session: %w", err)
	}

	return &CheckoutSession{
		SessionID:   sess.ID,
		CheckoutURL: sess.URL,
	}, nil
}

func (s *RealStripe) CancelSubscription(ctx context.Context, userID string) error {
	u, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if u.StripeCustomerID == "" {
		return fmt.Errorf("user has no Stripe customer")
	}

	// List active subscriptions for this customer and cancel the first one.
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(u.StripeCustomerID),
		Status:   stripe.String(string(stripe.SubscriptionStatusActive)),
	}
	iter := subscription.List(params)
	for iter.Next() {
		sub := iter.Subscription()
		_, err := subscription.Cancel(sub.ID, nil)
		if err != nil {
			return fmt.Errorf("stripe cancel subscription %s: %w", sub.ID, err)
		}
		return nil
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("stripe list subscriptions: %w", err)
	}
	return fmt.Errorf("no active subscription found")
}

func (s *RealStripe) CreatePortalSession(ctx context.Context, userID string) (*PortalSession, error) {
	customerID, err := s.ensureCustomer(ctx, userID)
	if err != nil {
		return nil, err
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(s.portalReturnURL),
	}
	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe create portal session: %w", err)
	}

	return &PortalSession{URL: sess.URL}, nil
}

func (s *RealStripe) GetSubscriptionDetails(ctx context.Context, userID string) (*SubscriptionDetails, error) {
	u, err := s.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if u.StripeCustomerID == "" {
		return nil, nil // No Stripe customer yet — free tier.
	}

	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(u.StripeCustomerID),
	}
	iter := subscription.List(params)
	for iter.Next() {
		sub := iter.Subscription()
		tier := tierFromMetadata(sub.Metadata)
		if tier == "" {
			tier = tierFromPriceID(sub, s.prices)
		}
		return &SubscriptionDetails{
			ID:                sub.ID,
			Status:            string(sub.Status),
			Tier:              tier,
			CurrentPeriodEnd:  time.Unix(sub.CurrentPeriodEnd, 0),
			CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		}, nil
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("stripe list subscriptions: %w", err)
	}
	return nil, nil // No subscription — free tier.
}

func (s *RealStripe) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(ctx, event.Data.Raw)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event.Data.Raw)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event.Data.Raw)
	case "invoice.payment_failed":
		return s.handlePaymentFailed(ctx, event.Data.Raw)
	default:
		log.Printf("[stripe] unhandled event type: %s", event.Type)
	}
	return nil
}

func (s *RealStripe) handleCheckoutCompleted(ctx context.Context, raw json.RawMessage) error {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return fmt.Errorf("unmarshal checkout session: %w", err)
	}

	customerID := sess.Customer.ID
	tier := sess.Metadata["relix_tier"]
	if tier == "" {
		log.Printf("[stripe] checkout.session.completed missing relix_tier metadata, customer=%s", customerID)
		return nil
	}

	return s.updateUserTier(ctx, customerID, user.Tier(tier))
}

func (s *RealStripe) handleSubscriptionUpdated(ctx context.Context, raw json.RawMessage) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(raw, &sub); err != nil {
		return fmt.Errorf("unmarshal subscription: %w", err)
	}

	customerID := sub.Customer.ID
	if sub.Status != stripe.SubscriptionStatusActive {
		return nil
	}

	tier := tierFromMetadata(sub.Metadata)
	if tier == "" {
		tier = tierFromPriceID(&sub, s.prices)
	}
	if tier == "" {
		log.Printf("[stripe] subscription.updated: could not determine tier for sub=%s", sub.ID)
		return nil
	}

	return s.updateUserTier(ctx, customerID, user.Tier(tier))
}

func (s *RealStripe) handleSubscriptionDeleted(ctx context.Context, raw json.RawMessage) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(raw, &sub); err != nil {
		return fmt.Errorf("unmarshal subscription: %w", err)
	}

	return s.updateUserTier(ctx, sub.Customer.ID, user.TierFree)
}

func (s *RealStripe) handlePaymentFailed(ctx context.Context, raw json.RawMessage) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(raw, &inv); err != nil {
		return fmt.Errorf("unmarshal invoice: %w", err)
	}
	log.Printf("[stripe] payment failed: customer=%s invoice=%s", inv.Customer.ID, inv.ID)
	return nil
}

func (s *RealStripe) updateUserTier(ctx context.Context, stripeCustomerID string, tier user.Tier) error {
	u, err := s.userStore.GetUserByStripeCustomerID(ctx, stripeCustomerID)
	if err != nil {
		return fmt.Errorf("find user by stripe customer %s: %w", stripeCustomerID, err)
	}
	u.Tier = tier
	if _, err := s.userStore.UpdateUser(ctx, u); err != nil {
		return fmt.Errorf("update user tier: %w", err)
	}
	log.Printf("[stripe] updated user %s to tier %s", u.ID, tier)
	return nil
}

// --- Helper functions ---

func tierFromMetadata(meta map[string]string) string {
	return meta["relix_tier"]
}

func tierFromPriceID(sub *stripe.Subscription, prices PriceConfig) string {
	if sub.Items == nil {
		return ""
	}
	for _, item := range sub.Items.Data {
		if item.Price == nil {
			continue
		}
		pid := item.Price.ID
		switch pid {
		case prices.PlusMonthly, prices.PlusAnnual:
			return string(user.TierPlus)
		case prices.ProMonthly, prices.ProAnnual:
			return string(user.TierPro)
		case prices.TeamMonthly, prices.TeamAnnual:
			return string(user.TierTeam)
		}
	}
	return ""
}

// --- StubStripe implementation ---

// StubStripe is a no-op StripeService that logs calls and returns success.
// Used when STRIPE_SECRET_KEY is not set.
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

func (s *StubStripe) CreatePortalSession(_ context.Context, userID string) (*PortalSession, error) {
	log.Printf("[stripe stub] CreatePortalSession: userID=%s", userID)
	return &PortalSession{
		URL: fmt.Sprintf("https://billing.stripe.com/stub/portal?user=%s", userID),
	}, nil
}

func (s *StubStripe) GetSubscriptionDetails(_ context.Context, userID string) (*SubscriptionDetails, error) {
	log.Printf("[stripe stub] GetSubscriptionDetails: userID=%s", userID)
	return nil, nil
}

func (s *StubStripe) HandleWebhook(_ context.Context, _ []byte, _ string) error {
	log.Printf("[stripe stub] HandleWebhook called")
	return nil
}

// --- Webhook HTTP handler helper ---

// WebhookHandler returns an http.HandlerFunc that reads the request body,
// passes it to the StripeService.HandleWebhook, and returns 200.
func WebhookHandler(svc StripeService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const maxBodySize = 65536
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
		if err != nil {
			http.Error(w, "read body failed", http.StatusBadRequest)
			return
		}

		sig := r.Header.Get("Stripe-Signature")
		if err := svc.HandleWebhook(r.Context(), body, sig); err != nil {
			log.Printf("[stripe] webhook error: %v", err)
			http.Error(w, "webhook processing failed", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
