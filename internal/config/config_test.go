package config_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/config"
)

func TestLoadConfig(t *testing.T) {
	config, err := config.NewConfig("./testdata/config")
	if err != nil {
		t.Fatalf("Error loading config file: %v", err)
	}

	pm := []struct {
		input int
		want  int
	}{
		{config.HTTPServer.Port, 9000},
		{config.DB.Port, 3306},
	}

	for _, pv := range pm {
		if pv.input != pv.want {
			t.Errorf("%d != %d\n", pv.input, pv.want)
		}
	}

	m := []struct {
		input string
		want  string
	}{
		{config.HTTPServer.Server, "example.com"},
		{config.DB.Server, "dbexample.com"},
		{config.DB.User, "root"},
		{config.DB.Pass, "root"},
		{config.DB.Name, "test"},
		{config.Crawler.Agent, "testing"},
	}

	for _, v := range m {
		if v.input != v.want {
			t.Errorf("%s != %s\n", v.input, v.want)
		}
	}
}
