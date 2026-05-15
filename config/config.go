package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	PG_DSN string `mapstructure:"PG_DSN"`
	PORT   string `mapstructure:"PORT"`
}

func LoadConfig() (config Config, err error) {
	viper.SetDefault("PORT", "8080")
	viper.AutomaticEnv()

	config.PORT = viper.GetString("PORT")
	config.PG_DSN = viper.GetString("PG_DSN")
	return
}
