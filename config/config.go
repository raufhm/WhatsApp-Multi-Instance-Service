package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	PGDSN    string
	Port     string
	S3Bucket string
}

func LoadConfig() (Config, error) {
	viper.SetDefault("PORT", "8080")
	viper.AutomaticEnv()

	return Config{
		Port:     viper.GetString("PORT"),
		PGDSN:    viper.GetString("PG_DSN"),
		S3Bucket: viper.GetString("S3_BUCKET"),
	}, nil
}
