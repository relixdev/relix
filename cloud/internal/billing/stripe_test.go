package billing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	stripe "github.com/stripe/stripe-go/v76"

	"github.com/relixdev/relix/cloud/internal/user"
)

func TestStubStripe_CreateCheckoutSession(t *testing.T) {
	stub := NewStubStripe()
	ctx := context.Background()

	sess, err := stub.CreateCheckoutSession(ctx, "usr_123", "plus")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.SessionID == "" {
		t.Error("expected non-empty session ID")
	}
	if sess.CheckoutURL == "" {
		t.Error("expected non-empty checkout URL")
	}
}

func TestStubStripe_CancelSubscription(t *testing.T) {
	stub := NewStubStripe()
	if err := stub.CancelSubscription(context.Background(), "usr_123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStubStripe_CreatePortalSession(t *testing.T) {
	stub := NewStubStripe()
	portal, err := stub.CreatePortalSession(context.Background(), "usr_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if portal.URL == "" {
		t.Error("expected non-empty portal URL")
	}
}

func TestStubStripe_GetSubscriptionDetails(t *testing.T) {
	stub := NewStubStripe()
	details, err := stub.GetSubscriptionDetails(context.Background(), "usr_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details != nil {
		t.Error("expected nil details from stub")
	}
}

func TestStubStripe_HandleWebhook(t *testing.T) {
	stub := NewStubStripe()
	if err := stub.HandleWebhook(context.Background(), []byte("{}"), "sig"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStubStripe_ImplementsInterface(t *testing.T) {
	var _ StripeService = (*StubStripe)(nil)
	var _ StripeService = (*RealStripe)(nil)
}

func TestPriceConfig_PriceForTier(t *testing.T) {
	pc := PriceConfig{
		PlusMonthly: "price_plus_m",
		PlusAnnual:  "price_plus_a",
		ProMonthly:  "price_pro_m",
		ProAnnual:   "price_pro_a",
		TeamMonthly: "price_team_m",
		TeamAnnual:  "price_team_a",
	}

	tests := []struct {
		tier    string
		want    string
		wantErr bool
	}{
		{"plus", "price_plus_m", false},
		{"pro", "price_pro_m", false},
		{"team", "price_team_m", false},
		{"free", "", true},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			got, err := pc.priceForTier(tt.tier)
			if (err != nil) != tt.wantErr {
				t.Errorf("priceForTier(%q) error = %v, wantErr %v", tt.tier, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("priceForTier(%q) = %q, want %q", tt.tier, got, tt.want)
			}
		})
	}
}

func TestPriceConfig_MissingPriceReturnsError(t *testing.T) {
	pc := PriceConfig{} // all empty

	for _, tier := range []string{"plus", "pro", "team"} {
		_, err := pc.priceForTier(tier)
		if err == nil {
			t.Errorf("expected error for tier %q with empty config", tier)
		}
	}
}

func TestTierFromPriceID(t *testing.T) {
	prices := PriceConfig{
		PlusMonthly: "price_plus_m",
		PlusAnnual:  "price_plus_a",
		ProMonthly:  "price_pro_m",
		ProAnnual:   "price_pro_a",
		TeamMonthly: "price_team_m",
		TeamAnnual:  "price_team_a",
	}

	tests := []struct {
		priceID string
		want    string
	}{
		{"price_plus_m", "plus"},
		{"price_plus_a", "plus"},
		{"price_pro_m", "pro"},
		{"price_pro_a", "pro"},
		{"price_team_m", "team"},
		{"price_team_a", "team"},
		{"price_unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.priceID, func(t *testing.T) {
			sub := &stripe.Subscription{
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{Price: &stripe.Price{ID: tt.priceID}},
					},
				},
			}
			got := tierFromPriceID(sub, prices)
			if got != tt.want {
				t.Errorf("tierFromPriceID(%q) = %q, want %q", tt.priceID, got, tt.want)
			}
		})
	}
}

func TestTierFromPriceID_NilItems(t *testing.T) {
	prices := PriceConfig{PlusMonthly: "price_plus_m"}
	sub := &stripe.Subscription{Items: nil}
	got := tierFromPriceID(sub, prices)
	if got != "" {
		t.Errorf("expected empty tier for nil items, got %q", got)
	}
}

func TestTierFromMetadata(t *testing.T) {
	tests := []struct {
		meta map[string]string
		want string
	}{
		{map[string]string{"relix_tier": "plus"}, "plus"},
		{map[string]string{"relix_tier": "pro"}, "pro"},
		{map[string]string{}, ""},
		{nil, ""},
	}

	for _, tt := range tests {
		got := tierFromMetadata(tt.meta)
		if got != tt.want {
			t.Errorf("tierFromMetadata(%v) = %q, want %q", tt.meta, got, tt.want)
		}
	}
}

func TestWebhookHandler_StubReturns200(t *testing.T) {
	stub := NewStubStripe()
	handler := WebhookHandler(stub)

	body := strings.NewReader(`{"type":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/billing/webhook", body)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestMemoryStore_StripeCustomerID(t *testing.T) {
	store := user.NewMemoryStore()
	ctx := context.Background()

	created, err := store.CreateUser(ctx, &user.User{Email: "test@example.com"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Fetch a fresh copy (as real code does via GetUserByID), then update.
	u, err := store.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	u.StripeCustomerID = "cus_test123"
	updated, err := store.UpdateUser(ctx, u)
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.StripeCustomerID != "cus_test123" {
		t.Errorf("expected StripeCustomerID=cus_test123, got %q", updated.StripeCustomerID)
	}

	// Look up by Stripe customer ID.
	found, err := store.GetUserByStripeCustomerID(ctx, "cus_test123")
	if err != nil {
		t.Fatalf("get by stripe customer: %v", err)
	}
	if found.ID != u.ID {
		t.Errorf("expected user ID %q, got %q", u.ID, found.ID)
	}

	// Non-existent Stripe customer ID.
	_, err = store.GetUserByStripeCustomerID(ctx, "cus_nonexistent")
	if err == nil {
		t.Error("expected error for non-existent Stripe customer ID")
	}
}

func TestMemoryStore_StripeCustomerID_UpdateIndex(t *testing.T) {
	store := user.NewMemoryStore()
	ctx := context.Background()

	u, err := store.CreateUser(ctx, &user.User{Email: "test@example.com"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Set initial customer ID.
	u.StripeCustomerID = "cus_old"
	if _, err := store.UpdateUser(ctx, u); err != nil {
		t.Fatalf("update user: %v", err)
	}

	// Change customer ID — old one should no longer resolve.
	u.StripeCustomerID = "cus_new"
	if _, err := store.UpdateUser(ctx, u); err != nil {
		t.Fatalf("update user: %v", err)
	}

	_, err = store.GetUserByStripeCustomerID(ctx, "cus_old")
	if err == nil {
		t.Error("expected error for old Stripe customer ID after update")
	}

	found, err := store.GetUserByStripeCustomerID(ctx, "cus_new")
	if err != nil {
		t.Fatalf("get by new stripe customer: %v", err)
	}
	if found.ID != u.ID {
		t.Errorf("expected user ID %q, got %q", u.ID, found.ID)
	}
}
