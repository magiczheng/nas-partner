package main

import (
	"fmt"
	"log"

	"nas-partner/backend/internal/config"
	"nas-partner/backend/internal/router"
)

func main() {
	cfg := config.Load()
	r := router.New(cfg)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
