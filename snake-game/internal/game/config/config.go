package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	proto "snake-game/internal/proto/gen"
)

func LoadConfig(path string) (*proto.GameConfig, error) {
	var cfg proto.GameConfig
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
