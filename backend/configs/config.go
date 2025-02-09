package configs

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type ConsumerConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	GroupID         string `yaml:"group_id"`
	AutoOffsetReset string `yaml:"auto_offset_reset"`
	PingTopic       string `yaml:"ping_topic"`
}

type ServerConfig struct {
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type DBConfig struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	SSLMode string `yaml:"ssl_mode"`
}

type Config struct {
	ConsumerConfig ConsumerConfig `yaml:"consumer"`
	ServerConfig   ServerConfig   `yaml:"server"`
	DBConfig       DBConfig       `yaml:"db"`
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
