package config

import (
	"github.com/stjudewashere/seonaut/internal/datastore"

	"github.com/spf13/viper"
)

type HTTPServerConfig struct {
	Server string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
}

type CrawlerConfig struct {
	Agent string `mapstructure:"agent"`
}

// Config stores the configuration for the application.
type Config struct {
	Crawler    CrawlerConfig       `mapstructure:"crawler"`
	HTTPServer HTTPServerConfig    `mapstructure:"server"`
	DB         *datastore.DBConfig `mapstructure:"database"`
}

// NewConfig loads the configuration from the specified file and path.
func NewConfig(path, filename string) (*Config, error) {
	viper.SetConfigName(filename)
	viper.SetConfigType("toml")
	viper.AddConfigPath(path)
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
