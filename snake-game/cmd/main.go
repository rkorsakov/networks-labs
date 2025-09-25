package main

import (
	"log"
	"os"
	"snake-game/internal/game/config"
	"snake-game/internal/game/core"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Print("CONFIG_PATH env variable not set")
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	game := core.NewGame(cfg)
	game.Start()
}
