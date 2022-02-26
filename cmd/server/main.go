package main

import (
	"github.com/mnlg/lenkrr/internal/app"
)

func main() {
	lenkrr := app.NewApp(".")
	lenkrr.Run()
}
