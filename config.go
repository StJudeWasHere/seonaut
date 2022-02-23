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

type StripeConfig struct {
	Key             string
	Secret          string
	WebhookSecret   string
	AdvancedPriceId string
	ReturnURL       string
}

type Config struct {
	Server       string
	ServerPort   int
	CrawlerAgent string

	DB     DBConfig
	Stripe StripeConfig
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

	config.Stripe.Key = viper.GetString("Stripe.key")
	config.Stripe.Secret = viper.GetString("Stripe.secret")
	config.Stripe.WebhookSecret = viper.GetString("Stripe.webhook_secret")
	config.Stripe.AdvancedPriceId = viper.GetString("Stripe.advanced_price_id")
	config.Stripe.ReturnURL = viper.GetString("Stripe.return_url")

	return &config, nil
}
