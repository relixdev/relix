package main

import (
	"fmt"
	"log"
	"net/http"

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

	store := user.NewMemoryStore()
	tokens := auth.NewTokenService(cfg.JWTSecret)
	reg := machine.NewRegistry(store)

	srv := api.New(api.Config{
		Tokens:      tokens,
		EmailAuth:   auth.NewEmailAuth(store),
		GitHubOAuth: auth.NewGitHubOAuth(cfg.GitHubClientID, cfg.GitHubClientSecret),
		UserStore:   store,
		Registry:    reg,
		Stripe:      billing.NewStubStripe(),
		Push:        push.NewAPNs(),
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("relix cloud listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server: %v", err)
	}
}
