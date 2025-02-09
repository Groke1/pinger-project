package configs

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type PingerConfig struct {
	ContainersTimeout time.Duration `yaml:"containers_timeout"`
	PingTimeout       time.Duration `yaml:"ping_timeout"`
	Workers           int           `yaml:"workers"`
}

type ProducerConfig struct {
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	PingTopic string `yaml:"ping_topic"`
}

type Config struct {
	ProducerConfig ProducerConfig `yaml:"producer"`
	PingerConfig   PingerConfig   `yaml:"server"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
