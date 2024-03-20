package main

import (
	"flag"

	"github.com/stjudewashere/seonaut/internal/container"
	"github.com/stjudewashere/seonaut/internal/http"
)

func main() {
	var configFile string

	flag.StringVar(&configFile, "c", "config", "Specify configuration file. Default is config.")
	flag.Parse()

	container := container.NewContainer(configFile)
	http.NewServer(container)
}
