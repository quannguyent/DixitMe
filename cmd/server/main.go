package main

import (
	"log"

	"dixitme/internal/app"
	"dixitme/internal/config"
)

func main() {
	cfg := config.Load()
	if err := app.Run(cfg); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
