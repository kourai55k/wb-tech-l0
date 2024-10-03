package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	HttpServer
	Kafka
	Database
	Env string
}

type HttpServer struct {
	Host string `yaml:"host" env-default:"localhost"`
	Port int    `yaml:"port" env-default:"8080"`
}

type Kafka struct {
	Brokers []string `yaml:"brokers" env-default:"{localhost:9093}"`
	Topic   string   `yaml:"topic" env-default:"orders"`
	GroupId string   `yaml:"groupId" env-default:"consumer-group"`
}

type Database struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5433"`
	User     string `yaml:"user" env-default:"kourai"`
	Password string `yaml:"password" env-default:"kourai123"`
	DBname   string `yaml:"dbname" env-default:"orders"`
}

func MustLoad() (*Config, error) {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	//Получаем путь до конфиг-файла из env-переменной CONFIG_PATH
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	// Проверяем существование конфиг-файла
	if _, err = os.Stat(configPath); err != nil {
		return nil, err
	}

	var cfg Config

	// Читаем конфиг-файл и заполняем нашу структуру
	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return nil, err
	}
	//log.Printf("Config loaded successfully: %+v", cfg)
	return &cfg, nil
}
