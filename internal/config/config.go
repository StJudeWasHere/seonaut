package config

import (
	"github.com/spf13/viper"
)

type HTTPServerConfig struct {
	Server string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
}

type DBConfig struct {
	Server string `mapstructure:"server"`
	Port   int    `mapstructure:"port"`
	User   string `mapstructure:"user"`
	Pass   string `mapstructure:"password"`
	Name   string `mapstructure:"database"`
}

type CrawlerConfig struct {
	Agent string `mapstructure:"agent"`
}

type Config struct {
	Crawler    CrawlerConfig    `mapstructure:"crawler"`
	HTTPServer HTTPServerConfig `mapstructure:"server"`
	DB         DBConfig         `mapstructure:"database"`
}

func NewConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
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
