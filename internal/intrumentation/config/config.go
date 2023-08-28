package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Env           string
	ListenAddress string
	LogLevel      string
}

func (c Config) IsDev() bool {
	return c.Env == "dev"
}

func Get() (*Config, error) {
	_ = viper.BindEnv("env")
	_ = viper.BindEnv("log_level")
	viper.SetEnvPrefix("ghoma")
	viper.AutomaticEnv()

	viper.SetDefault("listen_address", ":4196")
	viper.SetDefault("env", "prod")
	viper.SetDefault("log_level", "info")

	conf := &Config{}

	conf.Env = viper.GetString("env")
	conf.ListenAddress = viper.GetString("listen_address")
	if conf.Env == "dev" {
		viper.SetDefault("log_level", "debug")
	}
	conf.LogLevel = viper.GetString("log_level")

	return conf, nil
}
