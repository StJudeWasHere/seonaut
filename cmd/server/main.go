package main

import (
	"log"

	"github.com/mnlg/lenkrr/internal/app"
	"github.com/mnlg/lenkrr/internal/config"
)

func main() {
	config, err := config.NewConfig(".")
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	datastore, err := app.NewDataStore(config.DB)
	if err != nil {
		log.Fatalf("Error creating new datastore: %v\n", err)
	}

	lenkrr := app.NewApp(
		config,
		datastore,
	)

	lenkrr.Run()
}
