package main

import (
	"WBTechL0/internal/cache"
	"WBTechL0/internal/config"
	"WBTechL0/internal/consumer"
	"WBTechL0/internal/db"
	"WBTechL0/internal/db/repository"
	"WBTechL0/internal/http"
	"WBTechL0/internal/service"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Инициализируем логгер
	sl := setupLogger(envDev)
	sl.Debug("Logger initialized")

	// Загружаем конфиг
	sl.Info("Loading config")
	cfg, err := config.MustLoad()
	if err != nil {
		sl.Error("Error in loading config:", "error", err)
	}
	sl.Info("Config loaded successfully", "config", cfg)

	// Подключаемся к бд
	sl.Info("Connecting to database", "dbName", cfg.DBname)
	conn, err := db.ConnectToDB(cfg.Database)
	if err != nil {
		sl.Error("Failed to connect to db", "error", err)
		os.Exit(0)
	}
	sl.Info("Connect to database successfully")

	// Создаем таблицы
	sl.Info("Creating database tables")
	err = db.CreateTables(conn)
	if err != nil {
		sl.Error("Failed to create tables", "error", err)
	}
	sl.Info("Tables created successfully")

	// Инициализируем репозиторий
	sl.Info("Initializing repository")
	repo := repository.New(conn, sl)

	// Инициализируем кэш
	sl.Info("Initializing cache")
	cache1 := cache.New()
	// Восстанавливаем кэш из бд
	sl.Info("Restoring cache from db", "dbName", cfg.DBname)
	c, err := cache1.RestoreCacheFromDB(repo)
	if err != nil {
		sl.Error("Failed to init cache", "error", err)
	}
	sl.Info("Cache restored successfully", "Amount of restored items", c)

	// Инициализируем сервис
	sl.Info("Initializing order service")
	orderService := service.New(cache1, repo, sl)

	// Инициализируем сервер
	sl.Info("Initializing http server")
	httpServer := http.New(orderService, cfg.HttpServer)

	// Инициализируем коснюмер
	sl.Info("Initializing kafka consumer")
	kafkaConsumer := consumer.New(orderService, cfg.Kafka)

	// Запускаем
	// Запускаем HTTP-сервер в горутине
	go func() {
		sl.Info("Start http server", "host", cfg.HttpServer.Host, "port", cfg.HttpServer.Port)
		httpServer.Start()
	}()

	// Запускаем консюмер в горутине
	go func() {
		sl.Info("Start kafka consumer", "brokers", cfg.Kafka.Brokers, "topic", cfg.Topic)
		if err = kafkaConsumer.Start(context.Background()); err != nil {
			sl.Error("Failed to start Kafka consumer", err)
		}
	}()
	sl.Info("App is working")
	// Ожидаем сигнал завершения
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
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
