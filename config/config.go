package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	PrivateKey string
	Services   []ServiceConfig
	SparkleX   SparkleXConfig
	Networks   map[string]NetworkConfig
}

type ServiceConfig struct {
	From string
	To   string
}

type NetworkConfig struct {
	URL           string
	BridgeAddress string
	GEREAddress   string
}

type SparkleXConfig struct {
	URL                   string
	ReducerAddress        string
	NetworkManagerAddress string
	GERManagerAddress     string
}

func LoadConfig(name string, paths ...string) (Config, error) {
	config := Config{}
	v := viper.New()

	v.SetConfigName(name)
	v.SetConfigType("yaml")

	if len(paths) == 0 {
		paths = append(paths, ".")
	}

	for _, p := range paths {
		v.AddConfigPath(p)
	}

	err := v.ReadInConfig()
	if err != nil {
		return Config{}, err
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
