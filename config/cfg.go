package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	SrvPort string `mapstructure:"srv-port"`
	PgUser  string `mapstructure:"pg-user"`
	PgPswd  string `mapstructure:"pg-pswd"`
	PgDB    string `mapstructure:"pg-db"`
	PgHost  string `mapstructure:"pg-host"`
	PgPort  string `mapstructure:"pg-port"`
	ConnStr string `mapstructure:"conn-str"`
}

func LoadCfg() (*Config, error) {
	cfg := &Config{}

	filePath := os.Getenv("CONFIG_FILE_PATH")
	if filePath == "" {
		filePath = "/app/config/cfg.yaml"
	}

	viper.SetConfigFile(filePath)
	//viper.SetConfigFile("/Users/iGamez/Desktop/Cryptoproject-1/config/cfg.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return cfg, nil
}
