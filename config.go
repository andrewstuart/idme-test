package main

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	lg = logrus.New()
)

func setupConfig() *viper.Viper {
	cfg := viper.New()
	cfg.AddConfigPath(".")
	cfg.AddConfigPath("$HOME/idme")
	cfg.AddConfigPath("/etc/idme")

	cfg.SetConfigName("idme")
	cfg.SetEnvPrefix("IDME")

	cfg.AutomaticEnv()
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := cfg.ReadInConfig(); err != nil {
		lg.WithError(err).Error("could not read initial config")
	}

	cfg.OnConfigChange(func(_ fsnotify.Event) {
		if err := cfg.ReadInConfig(); err != nil {
			lg.WithError(err).Warn("could not reload config")
		}
	})

	go cfg.WatchConfig()

	return cfg
}
