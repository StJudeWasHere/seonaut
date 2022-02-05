package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server       string
	ServerPort   int
	DbServer     string
	DbPort       int
	DbUser       string
	DbPass       string
	DbName       string
	CrawlerAgent string
}

var config Config

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	config.Server = viper.GetString("Server.host")
	config.ServerPort = viper.GetInt("Server.port")
	config.DbServer = viper.GetString("Database.server")
	config.DbPort = viper.GetInt("Database.port")
	config.DbUser = viper.GetString("Database.user")
	config.DbPass = viper.GetString("Database.password")
	config.DbName = viper.GetString("Database.database")
	config.CrawlerAgent = viper.GetString("Crawler.agent")

	initDatabase(&config)
}

func main() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/resources/", http.StripPrefix("/resources", fileServer))
	http.Handle("/robots.txt", fileServer)
	http.Handle("/favicon.ico", fileServer)

	http.HandleFunc("/", requireAuth(serveHome))
	http.HandleFunc("/new-project", requireAuth(serveProjectAdd))
	http.HandleFunc("/crawl", requireAuth(serveCrawl))
	http.HandleFunc("/issues", requireAuth(serveIssues))
	http.HandleFunc("/issues/view", requireAuth(serveIssuesView))
	http.HandleFunc("/resources", requireAuth(serveResourcesView))
	http.HandleFunc("/signup", serveSignup)
	http.HandleFunc("/signin", serveSignin)
	http.HandleFunc("/signout", requireAuth(serveSignout))

	http.HandleFunc("/download", requireAuth(serveDownloadAll))

	fmt.Printf("Starting server at %s on port %d...\n", config.Server, config.ServerPort)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.Server, config.ServerPort), nil)
	if err != nil {
		log.Println(err)
	}
}
