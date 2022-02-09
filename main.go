package main

const (
	configPath = "."
)

func main() {
	app := NewApp(configPath)
	app.Run()
}
