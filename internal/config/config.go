package config

import (
	"path/filepath"

	"github.com/stjudewashere/seonaut/internal/crawler_service"
	"github.com/stjudewashere/seonaut/internal/datastore"

	"github.com/spf13/viper"
)

// HTTPServerConfig stores the configuration for the HTTP server.
// It is loaded from the config package.
type HTTPServerConfig struct {
	Server string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	URL    string `mapstructure:"url"`
}

// Config stores the configuration for the application.
type Config struct {
	Crawler    *crawler_service.Config `mapstructure:"crawler"`
	HTTPServer *HTTPServerConfig       `mapstructure:"server"`
	DB         *datastore.DBConfig     `mapstructure:"database"`
}

// NewConfig loads the configuration from the specified file and path.
func NewConfig(configFile string) (*Config, error) {
	viper.AddConfigPath(filepath.Dir(configFile))
	viper.SetConfigName(filepath.Base(configFile))
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
