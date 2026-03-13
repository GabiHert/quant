package main

import (
	"embed"
	"log"

	"quant/internal/infra"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	err := infra.Run(assets)
	if err != nil {
		log.Fatal(err)
	}
}
