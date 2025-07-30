package main

import (
	"fmt"
	"log"
	"referee/internal/config"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	fmt.Printf("Config loaded: %+v\n", cfg)
}
