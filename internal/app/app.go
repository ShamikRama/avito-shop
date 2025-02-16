package app

import (
	"avito-shop/internal/api"
	"avito-shop/internal/config"
	"avito-shop/internal/db"
	"avito-shop/internal/logger"
	"avito-shop/internal/repository"
	"avito-shop/internal/service"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func RunApp() {
	cfg := config.MustLoad()

	// Инициализация лог
	logs := logger.NewLogger()
	logs.Info("Logger initialized")

	// Подключение к бд
	db, err := storage.InitPostgres(cfg)
	if err != nil {
		logs.Error("Failed to connect to database", zap.Error(err))
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Инициализация репозиториев
	repoAuth := repository.NewAuthRepo(db, logs)
	repoUser := repository.NewUserRepo(db, logs)
	logs.Info("Repos initialized")

	// Инициализация сервисов
	servAuth := service.NewAuthService(repoAuth, logs)
	servUser := service.NewUserService(repoUser, logs)
	logs.Info("Services initialized")

	// Инициализация обработчиков
	handlers := api.NewApi(logs, servAuth, servUser)
	logs.Info("Handlers initialized")

	// Инициализация роутера
	router := handlers.InitRoutes()
	logs.Info("Routes initialized")

	//Инициализация сервер
	server := config.InitHttpServer(*cfg, router)
	logs.Info("Server initialized")

	go func() {
		logs.Info("Server staring")

		if err := config.RunServer(server); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	err = db.Close()
	if err != nil {
		log.Fatalf("Database is not closed")
	}
	logs.Info("Database closed")
}
