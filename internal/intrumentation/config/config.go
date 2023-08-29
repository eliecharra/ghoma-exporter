package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Env                string `mapstructure:"env"`
	ListenAddress      string `mapstructure:"listen_address"`
	GhomaListenAddress string `mapstructure:"ghoma_listen_address"`
	LogLevel           string `mapstructure:"log_level"`
}

func (c Config) IsDev() bool {
	return c.Env == "dev"
}

func Get() (*Config, error) {
	_ = viper.BindEnv("env")
	_ = viper.BindEnv("log_level")
	viper.SetEnvPrefix("ghoma")
	viper.AutomaticEnv()

	viper.SetDefault("ghoma_listen_address", ":4196")
	viper.SetDefault("listen_address", ":10005")
	viper.SetDefault("env", "prod")
	viper.SetDefault("log_level", "info")

	if viper.GetString("env") == "dev" {
		viper.SetDefault("log_level", "debug")
	}

	conf := &Config{}
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
