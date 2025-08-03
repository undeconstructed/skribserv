package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DBDSN      string `yaml:"dbdsn"`
	ListenAddr string `yaml:"listen_addr"`
}

func ReadConfig(paths ...string) (*Config, string, error) {
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
		}

		var config Config

		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, path, err
		}

		fillConfig(&config, path)

		return &config, path, nil
	}

	return nil, "", errors.New("no config found")
}

func fillConfig(config *Config, _ string) {
	if config.ListenAddr == "" {
		config.ListenAddr = ":8080"
	}
}
