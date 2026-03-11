package api

import (
	"encoding/json"
	"net/http"

	"github.com/relixdev/relix/cloud/internal/auth"
	"github.com/relixdev/relix/cloud/internal/billing"
)

func (s *Server) handleGetPlan(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	u, err := s.userStore.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	plan := billing.GetPlan(u.Tier)
	writeJSON(w, http.StatusOK, map[string]any{
		"tier":                plan.Tier,
		"display_name":        plan.DisplayName,
		"monthly_price_cents": plan.MonthlyPriceCents,
		"machine_limit":       plan.MachineLimit,
		"session_limit":       plan.SessionLimit,
	})
}

func (s *Server) handleBillingCheckout(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		Tier string `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Tier == "" {
		writeError(w, http.StatusBadRequest, "tier is required")
		return
	}

	session, err := s.stripe.CreateCheckoutSession(r.Context(), userID, req.Tier)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "checkout failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"session_id":   session.SessionID,
		"checkout_url": session.CheckoutURL,
	})
}
