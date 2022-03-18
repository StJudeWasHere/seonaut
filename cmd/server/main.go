package main

import (
	"flag"
	"log"

	"github.com/stjudewashere/seonaut/internal/config"
	"github.com/stjudewashere/seonaut/internal/datastore"
	"github.com/stjudewashere/seonaut/internal/http"
)

func main() {
	var fname string
	var path string

	flag.StringVar(&fname, "c", "config", "Specify config filename. Default is config")
	flag.StringVar(&path, "p", ".", "Specify config path. Default is current directory")
	flag.Parse()

	config, err := config.NewConfig(path, fname)
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
