package filter

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/logger"
)

const defName = "default"

type Config struct {
	Service map[string][]string
	Client  map[string]map[string][]string
}

func loadConfig() *Config {
	serviceDefaultKey := "filters.service.default"
	if !config.Conf.GetViper().IsSet(serviceDefaultKey) {
		config.Conf.GetViper().Set(serviceDefaultKey, []string{"base"})
	}
	clientDefaultKey := "filters.client.default.default"
	if !config.Conf.GetViper().IsSet(clientDefaultKey) {
		config.Conf.GetViper().Set(clientDefaultKey, []string{"base"})
	}

	conf := &Config{
		Service: make(map[string][]string),
		Client:  make(map[string]map[string][]string),
	}
	err := config.Conf.Parse("filters", conf, true)
	if err != nil {
		logger.Log.Fatal("parse filter config err", zap.Error(err))
	}
	return conf
}
