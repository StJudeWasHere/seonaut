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

	if config.DbServer != "dbexample.com" {
		t.Error("DbServer != dbexample.com")
	}

	if config.DbPort != 3306 {
		t.Error("DbPort != 3306")
	}

	if config.DbUser != "root" {
		t.Error("DbUser != root")
	}

	if config.DbPass != "root" {
		t.Error("DbPass != root")
	}

	if config.DbName != "test" {
		t.Error("DbName != test")
	}

	if config.CrawlerAgent != "testing" {
		t.Error("CrawlerAgent != testing")
	}
}
