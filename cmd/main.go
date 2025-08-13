package main

import (
	"Effective_Mobile/internal/config"
	"Effective_Mobile/internal/repository"
	"Effective_Mobile/internal/router"
	"Effective_Mobile/internal/router/handlers"
	"Effective_Mobile/internal/service"
	"Effective_Mobile/pkg/logger"
	"go.uber.org/zap"
)

// @title Effective Mobile Subscription Service API
// @version 1.0
// @description API для управления подписками пользователей

// @contact.name API Support
// @contact.email ravilkarimov06@mail.ru

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-20.0.html

// @host localhost:8080
// @BasePath /

// SubscriptionHandler handles HTTP requests related to user subscriptions.
// It acts as the entry point for API requests, validating input, calling the service layer,
// and sending appropriate HTTP responses.
func main() {

	cfg := config.MustLoad()

	log, err := logger.NewLogger(cfg.LogLevel)

	if err != nil {
		panic(err)
	}

	storage, err := repository.NewStorage(cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode, log)
	if err != nil {
		log.Fatal("Error initializing storage")
	}
	defer storage.Close()

	repo := storage.NewRepository()

	subService := service.NewSubscriptionService(repo, log)

	handler := handlers.NewSubscriptionHandler(subService)
	log.Info("addr", zap.String("addr", cfg.Addr))
	rout := router.NewRouter(handler, log)
	if err := rout.RunRouter(cfg.Addr, cfg.RequestPerSecond, cfg.Burst); err != nil {
		log.Fatal("Error initializing router")
	}
}
