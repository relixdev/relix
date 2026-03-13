package api

import (
	"encoding/json"
	"net/http"

	"github.com/relixdev/relix/cloud/internal/auth"
	"github.com/relixdev/relix/cloud/internal/billing"
	"github.com/relixdev/relix/cloud/internal/machine"
	"github.com/relixdev/relix/cloud/internal/push"
	"github.com/relixdev/relix/cloud/internal/user"
)

// Server wires together all HTTP handlers and dependencies.
type Server struct {
	mux         *http.ServeMux
	tokens      *auth.TokenService
	emailAuth   *auth.EmailAuth
	githubOAuth *auth.GitHubOAuth
	userStore   user.Store
	registry    *machine.Registry
	stripe      billing.StripeService
	push        push.Service
}

// Config holds the dependencies needed to build a Server.
type Config struct {
	Tokens      *auth.TokenService
	EmailAuth   *auth.EmailAuth
	GitHubOAuth *auth.GitHubOAuth
	UserStore   user.Store
	Registry    *machine.Registry
	Stripe      billing.StripeService
	Push        push.Service
}

// New creates and configures a Server with all routes registered.
func New(cfg Config) *Server {
	s := &Server{
		mux:         http.NewServeMux(),
		tokens:      cfg.Tokens,
		emailAuth:   cfg.EmailAuth,
		githubOAuth: cfg.GitHubOAuth,
		userStore:   cfg.UserStore,
		registry:    cfg.Registry,
		stripe:      cfg.Stripe,
		push:        cfg.Push,
	}
	s.routes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	authMW := auth.Middleware(s.tokens)

	// Public auth routes
	s.mux.HandleFunc("POST /auth/github", s.handleAuthGitHub)
	s.mux.HandleFunc("POST /auth/email/register", s.handleEmailRegister)
	s.mux.HandleFunc("POST /auth/email/login", s.handleEmailLogin)
	s.mux.HandleFunc("POST /auth/refresh", authMW(http.HandlerFunc(s.handleAuthRefresh)).ServeHTTP)

	// Machine routes (authenticated)
	s.mux.Handle("GET /machines", authMW(http.HandlerFunc(s.handleListMachines)))
	s.mux.Handle("POST /machines", authMW(http.HandlerFunc(s.handleRegisterMachine)))
	s.mux.Handle("DELETE /machines/{id}", authMW(http.HandlerFunc(s.handleDeleteMachine)))

	// Push routes (authenticated)
	s.mux.Handle("POST /push/register", authMW(http.HandlerFunc(s.handlePushRegister)))
	s.mux.Handle("POST /push/send", authMW(http.HandlerFunc(s.handlePushSend)))

	// Billing routes (authenticated)
	s.mux.Handle("GET /billing/plan", authMW(http.HandlerFunc(s.handleGetPlan)))
	s.mux.Handle("POST /billing/checkout", authMW(http.HandlerFunc(s.handleBillingCheckout)))
	s.mux.Handle("POST /billing/portal", authMW(http.HandlerFunc(s.handleBillingPortal)))
	s.mux.Handle("GET /billing/subscription", authMW(http.HandlerFunc(s.handleBillingSubscription)))

	// Stripe webhook (public — verified by signature)
	s.mux.HandleFunc("POST /billing/webhook", billing.WebhookHandler(s.stripe).ServeHTTP)

	// Health check
	s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
}

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
