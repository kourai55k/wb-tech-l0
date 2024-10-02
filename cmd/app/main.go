package main

import (
	"WBTechL0/internal/cache"
	"WBTechL0/internal/config"
	"WBTechL0/internal/db"
	"WBTechL0/internal/db/repository"
	"WBTechL0/internal/service"
	"log"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Загружаем конфиг
	cfg := config.MustLoad()
	log.Printf("Config loaded succesfully: %+v", cfg)

	// Инициализируем логгер
	sl := setupLogger(cfg.Env)
	sl.Debug("Logger initialized")

	// Подключаемся к бд
	conn, err := db.ConnectToDB(cfg.Database)
	if err != nil {
		sl.Error("Failed to connect to db", err)
	}
	sl.Debug("Connect to database successfully")

	// Создаем таблицы
	err = db.CreateTables(conn)
	if err != nil {
		sl.Error("Failed to create tables", err)
	}
	sl.Debug("Tables created successfully")

	// Инициализируем репозиторий
	repo := repository.New(conn, sl)

	// Инициализируем кэш
	cache1 := cache.New()
	err = cache1.RestoreCacheFromDB(repo)
	if err != nil {
		sl.Error("Failed to init cache")
	}

	// Инициализируем сервис
	orderService := service.New(cache1, repo, sl)
	sl.Info("service", orderService)
}

func setupLogger(env string) *slog.Logger {
	var sl *slog.Logger

	switch env {
	case envLocal:
		sl = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		sl = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		sl = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return sl
}
