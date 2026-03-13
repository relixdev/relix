package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/relixdev/relix/cloud/internal/api"
	"github.com/relixdev/relix/cloud/internal/auth"
	"github.com/relixdev/relix/cloud/internal/billing"
	"github.com/relixdev/relix/cloud/internal/config"
	"github.com/relixdev/relix/cloud/internal/machine"
	"github.com/relixdev/relix/cloud/internal/push"
	"github.com/relixdev/relix/cloud/internal/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	var store user.Store
	if cfg.DatabaseURL != "" {
		db, err := sql.Open("postgres", cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("postgres open: %v", err)
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			log.Fatalf("postgres ping: %v", err)
		}
		if err := user.Migrate(db); err != nil {
			log.Fatalf("postgres migrate: %v", err)
		}
		store = user.NewPostgresStore(db)
		log.Printf("using PostgreSQL user store")
	} else {
		store = user.NewMemoryStore()
		log.Printf("using in-memory user store (set DATABASE_URL for PostgreSQL)")
	}

	tokens := auth.NewTokenService(cfg.JWTSecret)
	reg := machine.NewRegistry(store)

	var stripeSvc billing.StripeService
	if cfg.StripeSecretKey != "" {
		stripeSvc = billing.NewRealStripe(billing.RealStripeConfig{
			SecretKey:     cfg.StripeSecretKey,
			WebhookSecret: cfg.StripeWebhookSecret,
			UserStore:     store,
			Prices:        billing.LoadPriceConfig(),
		})
		log.Printf("using real Stripe billing")
	} else {
		stripeSvc = billing.NewStubStripe()
		log.Printf("using stub Stripe billing (set STRIPE_SECRET_KEY for real Stripe)")
	}

	srv := api.New(api.Config{
		Tokens:      tokens,
		EmailAuth:   auth.NewEmailAuth(store),
		GitHubOAuth: auth.NewGitHubOAuth(cfg.GitHubClientID, cfg.GitHubClientSecret),
		UserStore:   store,
		Registry:    reg,
		Stripe:      stripeSvc,
		Push:        push.NewAPNs(),
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("relix cloud listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server: %v", err)
	}
}
