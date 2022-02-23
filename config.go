package main

import (
	"github.com/spf13/viper"
)

type DBConfig struct {
	Server string
	Port   int
	User   string
	Pass   string
	Name   string
}

type Config struct {
	Server     string
	ServerPort int

	DB DBConfig

	CrawlerAgent string

	StripeKey            string
	StripeSecret         string
	StripeWebhookSecret  string
	StripeAdvancePriceId string
	StripeReturnURL      string
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

	config.DB.Server = viper.GetString("Database.server")
	config.DB.Port = viper.GetInt("Database.port")
	config.DB.User = viper.GetString("Database.user")
	config.DB.Pass = viper.GetString("Database.password")
	config.DB.Name = viper.GetString("Database.database")

	config.CrawlerAgent = viper.GetString("Crawler.agent")

	config.StripeKey = viper.GetString("Stripe.key")
	config.StripeSecret = viper.GetString("Stripe.secret")
	config.StripeWebhookSecret = viper.GetString("Stripe.webhook_secret")
	config.StripeAdvancePriceId = viper.GetString("Stripe.advanced_price_id")
	config.StripeReturnURL = viper.GetString("Stripe.return_url")

	return &config, nil
}
