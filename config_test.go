package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := loadConfig("./test/")
	if err != nil {
		t.Errorf("Error loading config file: %v\n", err)
	}

	if config.Server != "example.com" {
		t.Error("Server != example.com")
	}

	if config.ServerPort != 9000 {
		t.Error("ServerPort != 9000")
	}

	if config.DB.Server != "dbexample.com" {
		t.Error("DB.Server != dbexample.com")
	}

	if config.DB.Port != 3306 {
		t.Error("DB.Port != 3306")
	}

	if config.DB.User != "root" {
		t.Error("DB.User != root")
	}

	if config.DB.Pass != "root" {
		t.Error("DB.Pass != root")
	}

	if config.DB.Name != "test" {
		t.Error("DB.Name != test")
	}

	if config.CrawlerAgent != "testing" {
		t.Error("CrawlerAgent != testing")
	}
}
