package config

import (
	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/datastore"
	"github.com/stjudewashere/seonaut/internal/http"

	"github.com/spf13/viper"
)

// Config stores the configuration for the application.
type Config struct {
	Crawler    *crawler.CrawlerConfig `mapstructure:"crawler"`
	HTTPServer *http.HTTPServerConfig `mapstructure:"server"`
	DB         *datastore.DBConfig    `mapstructure:"database"`
}

// NewConfig loads the configuration from the specified file and path.
func NewConfig(path, filename string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(filename)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
