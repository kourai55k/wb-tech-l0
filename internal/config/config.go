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
	Brokers       []string `yaml:"brokers" env-default:"localhost:9093"`
	Topic         string   `yaml:"topic" env-default:"orders"`
	ConsumerGroup string   `yaml:"consumerGroup" env-default:"consumer-group"`
}

type Database struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5433"`
	User     string `yaml:"user" env-default:"kourai"`
	Password string `yaml:"password" env-default:"kourai123"`
	DBname   string `yaml:"dbname" env-default:"orders"`
}

func MustLoad() *Config {
	// Загружаем переменные окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Получаем путь до конфиг-файла из env-переменной CONFIG_PATH
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	// Проверяем существование конфиг-файла
	if _, err = os.Stat(configPath); err != nil {
		log.Fatalf("error opening config file: %s", err)
	}

	var cfg Config

	// Читаем конфиг-файл и заполняем нашу структуру
	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error reading config file: %s", err)
	}
	//log.Printf("Config loaded successfully: %+v", cfg)
	return &cfg
}
