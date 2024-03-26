package main

import (
	"flag"

	"github.com/stjudewashere/seonaut/internal/routes"
	"github.com/stjudewashere/seonaut/internal/services"
)

func main() {
	var configFile string

	flag.StringVar(&configFile, "c", "config", "Specify configuration file. Default is config.")
	flag.Parse()

	container := services.NewContainer(configFile)
	routes.NewServer(container)
}
