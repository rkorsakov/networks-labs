package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Field FieldConfig
}

type FieldConfig struct {
	Width  int `yaml:"width" default:"15"`
	Height int `yaml:"height" default:"15"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
