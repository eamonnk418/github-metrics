package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type Application struct {
	Config *Config
	Logger *slog.Logger
}

type Config struct {
	AppID          int64  `mapstructure:"APP_ID"`
	InstallationID int64  `mapstructure:"APP_INSTALLATION_ID"`
	PrivateKey     string `mapstructure:"APP_PRIVATE_KEY"`
	AccessToken    string `mapstructure:"ACCESS_TOKEN"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	viper.MustBindEnv("APP_ID")
	viper.MustBindEnv("APP_INSTALLATION_ID")
	viper.MustBindEnv("APP_PRIVATE_KEY")
	viper.MustBindEnv("ACCESS_TOKEN")

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
