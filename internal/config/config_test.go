package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := NewConfig("../../test/")
	if err != nil {
		t.Errorf("Error loading config file: %v\n", err)
	}

	if config.ServerPort != 9000 {
		t.Error("ServerPort != 9000")
	}

	if config.DB.Port != 3306 {
		t.Error("DB.Port != 3306")
	}

	m := []struct {
		input string
		want  string
	}{
		{config.Server, "example.com"},
		{config.DB.Server, "dbexample.com"},
		{config.DB.User, "root"},
		{config.DB.Pass, "root"},
		{config.DB.Name, "test"},
		{config.CrawlerAgent, "testing"},
		{config.Stripe.Key, "stripe_key"},
		{config.Stripe.Secret, "stripe_secret"},
		{config.Stripe.WebhookSecret, "webhook_secret"},
		{config.Stripe.AdvancedPriceId, "price_id"},
		{config.Stripe.ReturnURL, "http://localhost:9000"},
	}

	for _, v := range m {
		if v.input != v.want {
			t.Errorf("%s != %s\n", v.input, v.want)
		}
	}
}
