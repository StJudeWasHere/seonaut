package main

import (
	"log"
)

var config *Config

func main() {
	var err error
	config, err = loadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	initDatabase(config)
	initServer()
}
