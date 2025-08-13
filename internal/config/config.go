package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	Storage
	Rest
	RateLimit
	LogLevel string `yaml:"log_level"`
}
type Storage struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"ssl_mode"`
}
type Rest struct {
	Addr string `yaml:"addr"`
}
type RateLimit struct {
	RequestPerSecond int `yaml:"request_per_second"`
	Burst            int `yaml:"burst"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.yaml" // Путь по умолчанию
	}

	// Преобразуем относительный путь в абсолютный
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to get absolute path: %w", err))
	}

	configFile, err := os.Open(absPath)
	if err != nil {
		panic(fmt.Errorf("failed to open config file at %s: %w", absPath, err))
	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		panic(fmt.Errorf("failed to decode config: %w", err))
	}

	return &config
}
