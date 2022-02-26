package app

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := NewConfig("../../test/")
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

	if config.Stripe.Key != "stripe_key" {
		t.Error("Stripe.Key != stripe_key")
	}

	if config.Stripe.Secret != "stripe_secret" {
		t.Error("Stripe.Secret != stripe_secret")
	}

	if config.Stripe.WebhookSecret != "webhook_secret" {
		t.Error("Stripe.WebhookSecret != webhook_secret")
	}

	if config.Stripe.AdvancedPriceId != "price_id" {
		t.Error("Stripe.AdvancedPriceId != price_id")
	}

	if config.Stripe.ReturnURL != "http://localhost:9000" {
		t.Error("Stripe.ReturnURL != http://localhost:9000")
	}
}
