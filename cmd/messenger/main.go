package main

import (
	"log"
	"messenger/internal/config"
	"messenger/internal/handlers"
	"messenger/internal/middleware"
	"messenger/internal/repository"
	"messenger/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключение к БД
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURI), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Автомиграция
	if err := db.AutoMigrate(&repository.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)

	// Инициализация сервисов
	authService := service.NewAuthService(userRepo, cfg)

	// Инициализация хендлеров
	authHandler := handlers.NewAuthHandler(authService)

	// Инициализация middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Настройка Gin
	r := gin.Default()

	// Публичные маршруты (без авторизации)
	api := r.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
	}

	// Защищённые маршруты (требуют авторизации)
	protected := api.Group("/")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.GET("/users/me", authHandler.GetMe)
		// Здесь будут другие защищённые эндпоинты (чаты, сообщения и т.д.)
	}

	// Запуск сервера
	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
