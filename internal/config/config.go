package config

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// CrawlerConfig stores the configuration for the crawler.
type CrawlerConfig struct {
	Agent string `mapstructure:"agent"`
}

// HTTPServerConfig stores the configuration for the HTTP server.
type HTTPServerConfig struct {
	Server string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	URL    string `mapstructure:"url"`
}

// DBConfig stores the configuration for the database store.
type DBConfig struct {
	Server string `mapstructure:"server"`
	Port   int    `mapstructure:"port"`
	User   string `mapstructure:"user"`
	Pass   string `mapstructure:"password"`
	Name   string `mapstructure:"database"`
}

// UIConfig stores the UI settings.
type UIConfig struct {
	Language string `mapstructure:"language"`
	Theme    string `mapstructure:"theme"`
}

// Config stores the configuration for the application.
type Config struct {
	Crawler    *CrawlerConfig    `mapstructure:"crawler"`
	HTTPServer *HTTPServerConfig `mapstructure:"server"`
	DB         *DBConfig         `mapstructure:"database"`
	UIConfig   *UIConfig         `mapstructure:"UI"`
}

// NewConfig loads the configuration from the specified file and path.
func NewConfig(configFile string) (*Config, error) {
	viper.AddConfigPath(filepath.Dir(configFile))
	viper.SetConfigName(filepath.Base(configFile))
	viper.SetConfigType("toml")

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SEONAUT") // Optional: prefix for env vars
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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
