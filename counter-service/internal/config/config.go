package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

//Config конфиг
type Config struct {
	Service Service `yaml:"service"`
	Redis   Redis   `yaml:"redis"`
	MQ      MQ      `yaml:"mq"`
}

//Service конфиг сервиса
type Service struct {
	Port        string `yaml:"port"`
	ServiceName string `yaml:"service_name"`
}

// Redis конфиг редиса
type Redis struct {
	Address string `yaml:"address"`
}

// MQ конфиг очереди сообщений
type MQ struct {
	ConnStr string `yaml:"conn_str"`
}

//ReadConfig читает конфиг из файла
func ReadConfig(path string) (Config, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	return cfg, err
}
