package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

//Config конфиг
type Config struct {
	Server           Server           `yaml:"server"`
	DialogueDatabase DialogueDatabase `yaml:"dialogue_database"`
}

//Server конфиг HTTP-сервера
type Server struct {
	Port string `yaml:"port"`
}

//DialogueDatabase конфиг БД диалогов
type DialogueDatabase struct {
	User     string   `yaml:"user"`
	Password string   `yaml:"pass"`
	DbName   string   `yaml:"db_name"`
	Shards   []string `yaml:"shards"`
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