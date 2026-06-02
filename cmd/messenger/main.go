package main

import (
	"log"

	"github.com/DNA-Z/Ignis/internal/config"
	"github.com/DNA-Z/Ignis/internal/handlers"
	"github.com/DNA-Z/Ignis/internal/middleware"
	"github.com/DNA-Z/Ignis/internal/repository"
	"github.com/DNA-Z/Ignis/internal/service"

	"github.com/DNA-Z/Ignis/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURI), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Chat{},
		&models.ChatParticipant{},
		&models.Message{},
		&models.ReadReceipt{},
		&models.File{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	participantRepo := repository.NewParticipantRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	fileRepo := repository.NewFileRepository(db)

	authService := service.NewAuthService(userRepo, cfg)
	chatService := service.NewChatService(chatRepo, participantRepo, userRepo, messageRepo)
	messageService := service.NewMessageService(messageRepo, participantRepo, userRepo, fileRepo)
	fileService := service.NewFileService(fileRepo, messageRepo, participantRepo, "./uploads", 10<<20) // 10MB
	searchService := service.NewSearchService(db, userRepo, chatRepo, messageRepo, participantRepo)

	authHandler := handlers.NewAuthHandler(authService)
	chatHandler := handlers.NewChatHandler(chatService)
	messageHandler := handlers.NewMessageHandler(messageService)
	fileHandler := handlers.NewFileHandler(fileService, 10<<20)
	searchHandler := handlers.NewSearchHandler(searchService)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
	}

	protected := api.Group("/")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Пользователи
		protected.GET("/users/me", authHandler.GetMe)

		// Чаты
		protected.POST("/chats", chatHandler.CreateChat)
		protected.GET("/chats", chatHandler.GetUserChats)
		protected.GET("/chats/:id", chatHandler.GetChat)
		protected.PUT("/chats/:id", chatHandler.UpdateChat)
		protected.DELETE("/chats/:id", chatHandler.DeleteChat)

		// Участники
		protected.POST("/chats/:id/users", chatHandler.AddParticipants)
		protected.DELETE("/chats/:id/users/:userId", chatHandler.RemoveParticipant)
		protected.PUT("/chats/:id/users/:userId/role", chatHandler.UpdateRole)

		// Сообщения
		protected.POST("/chats/:id/messages", messageHandler.SendMessage)
		protected.GET("/chats/:id/messages", messageHandler.GetMessages)
		protected.PUT("/messages/:id", messageHandler.UpdateMessage)
		protected.DELETE("/messages/:id", messageHandler.DeleteMessage)
		protected.POST("/messages/:id/read", messageHandler.MarkAsRead)

		// Файлы
		protected.POST("/messages/:id/files", fileHandler.UploadFile)
		protected.GET("/files/:id", fileHandler.GetFile)
		protected.DELETE("/files/:id", fileHandler.DeleteFile)

		// Поиск
		protected.GET("/search/messages", searchHandler.SearchMessages)
		protected.GET("/search/chats", searchHandler.SearchChats)
	}

	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
