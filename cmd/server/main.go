package main

import (
	"flag"
	"log"

	"github.com/edgekit/edgekit/internal/app"
	"github.com/edgekit/edgekit/internal/app/config"
)

func main() {
	configPath := flag.String("config", "configs/config.local.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	a := app.New(cfg)
	if err := a.Run(); err != nil {
		log.Fatalf("app run: %v", err)
	}
}
