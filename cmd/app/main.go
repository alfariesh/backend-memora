package main

import (
	"log"
	_ "time/tzdata"

	"github.com/alfariesh/backend-memora/config"
	"github.com/alfariesh/backend-memora/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
