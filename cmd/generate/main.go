package main

import (
	"log"

	"github.com/duynhlab/grafana-dashboards/internal/generate"
)

func main() {
	if err := generate.Run(); err != nil {
		log.Fatalf("generate: %v", err)
	}
}
