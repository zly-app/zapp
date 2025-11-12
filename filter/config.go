package filter

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/log"
)

const defName = "default"

type Config struct {
	Service map[string][]string
	Client  map[string]map[string][]string
}

func loadConfig() *Config {
	conf := &Config{
		Service: make(map[string][]string),
		Client:  make(map[string]map[string][]string),
	}
	err := config.Conf.Parse("filters", conf, true)
	if err != nil {
		log.Log.Fatal("parse filter config err", zap.Error(err))
	}
	return conf
}
