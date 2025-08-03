package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Cfg CfgStruct `yaml:"cfg"`
}

type CfgStruct struct {
	SrvPort string `yaml:"srv-port"`
	PgUser  string `yaml:"pg-user"`
	PgPswd  string `yaml:"pg-pswd"`
	PgDB    string `yaml:"pg-db"`
	PgHost  string `yaml:"pg-host"`
	PgPort  string `yaml:"pg-port"`
	URL     string `yaml:"url"`
}

func LoadCfg(filePath string) (*Config, error) {
	rawData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(rawData, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &cfg, nil
}

// package config

// import (
// 	"fmt"

// 	"github.com/spf13/viper"
// )

// type Config struct {
// 	SrvPort string `mapstructure:"srv-port"`
// 	PgUser  string `mapstructure:"pg-user"`
// 	PgPswd  string `mapstructure:"pg-pswd"`
// 	PgDB    string `mapstructure:"pg-db"`
// 	PgHost  string `mapstructure:"pg-host"`
// 	PgPort  string `mapstructure:"pg-port"`
// 	URL     string `mapstructure:"url"`
// }

// func LoadCfg() (*Config, error) {
// 	cfg := &Config{}

// 	viper.SetConfigFile("/Users/iGamez/Desktop/Cryptoproject-1/config/cfg.yaml")

// 	err := viper.ReadInConfig()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read configuration file: %w", err)
// 	}

// 	err = viper.Unmarshal(cfg)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
// 	}

// 	return cfg, nil
// }
