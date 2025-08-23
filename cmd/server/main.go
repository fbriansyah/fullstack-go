package main

import (
	"log"

	"go-templ-template/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	log.Printf("Starting server on port %s", cfg.Server.Port)
	log.Println("Configuration loaded successfully")
}
