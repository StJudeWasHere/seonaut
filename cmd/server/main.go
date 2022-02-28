package main

import (
	"log"

	"github.com/mnlg/lenkrr/internal/config"
	"github.com/mnlg/lenkrr/internal/datastore"
	"github.com/mnlg/lenkrr/internal/http"
)

func main() {
	config, err := config.NewConfig(".")
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	datastore, err := datastore.NewDataStore(config.DB)
	if err != nil {
		log.Fatalf("Error creating new datastore: %v\n", err)
	}

	lenkrr := http.NewApp(
		config,
		datastore,
	)

	lenkrr.Run()
}
