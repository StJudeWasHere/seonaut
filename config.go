package main

import (
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

func loadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config

	config.Server = viper.GetString("Server.host")
	config.ServerPort = viper.GetInt("Server.port")
	config.DbServer = viper.GetString("Database.server")
	config.DbPort = viper.GetInt("Database.port")
	config.DbUser = viper.GetString("Database.user")
	config.DbPass = viper.GetString("Database.password")
	config.DbName = viper.GetString("Database.database")
	config.CrawlerAgent = viper.GetString("Crawler.agent")

	return &config, nil
}
