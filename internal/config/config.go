package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

//Config конфиг
type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
}

//Server конфиг HTTP-сервера
type Server struct {
	Port string `yaml:"port"`
}

//Database конфиг БД
type Database struct {
	User     string   `yaml:"user"`
	Password string   `yaml:"pass"`
	Master   string   `yaml:"master"`
	Replicas []string `yaml:"replicas"`
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
