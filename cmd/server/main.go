package main

import (
	"io/ioutil"
	"log"

	"github.com/mnlg/lenkrr/internal/app"
	"github.com/mnlg/lenkrr/internal/user"

	"gopkg.in/yaml.v3"
)

func main() {
	config, err := app.NewConfig(".")
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	datastore, err := app.NewDataStore(config.DB)
	if err != nil {
		log.Fatalf("Error creating new datastore: %v\n", err)
	}

	translation, err := ioutil.ReadFile("translation.en.yaml")
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(translation, &m)
	if err != nil {
		log.Fatal(err)
	}

	userService := user.NewService(datastore)
	renderer := app.NewRenderer(m)

	lenkrr := app.NewApp(config, datastore, userService, renderer)
	lenkrr.Run()
}
