package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ListenAddress string
}

func Get() (*Config, error) {
	_ = viper.BindEnv("log_level")
	viper.SetEnvPrefix("ghoma")
	viper.AutomaticEnv()

	viper.SetDefault("listen_address", ":4196")

	conf := &Config{}

	conf.ListenAddress = viper.GetString("listen_address")

	return conf, nil
}
