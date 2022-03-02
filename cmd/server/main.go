package main

import (
	"log"

	"github.com/mnlg/seonaut/internal/config"
	"github.com/mnlg/seonaut/internal/datastore"
	"github.com/mnlg/seonaut/internal/http"
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

	err = datastore.Migrate()
	if err != nil {
		log.Fatalf("Error running migrations: %v\n", err)
	}

	server := http.NewApp(
		config,
		datastore,
	)

	server.Run()
}
