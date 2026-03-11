package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/relixdev/relix/relay/internal/config"
	"github.com/relixdev/relix/relay/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	srv := server.New()
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("relay listening on %s", addr)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
