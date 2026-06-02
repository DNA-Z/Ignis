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

	http := api.Group("/")
	http.Use(authMiddleware.RequireAuth())
	{
		// Пользователи
		http.GET("/users/me", authHandler.GetMe)

		// Чаты
		http.POST("/chats", chatHandler.CreateChat)
		http.GET("/chats", chatHandler.GetUserChats)
		http.GET("/chats/:id", chatHandler.GetChat)
		http.PUT("/chats/:id", chatHandler.UpdateChat)
		http.DELETE("/chats/:id", chatHandler.DeleteChat)

		// Участники
		http.POST("/chats/:id/users", chatHandler.AddParticipants)
		http.DELETE("/chats/:id/users/:userId", chatHandler.RemoveParticipant)
		http.PUT("/chats/:id/users/:userId/role", chatHandler.UpdateRole)

		// Сообщения
		http.POST("/chats/:id/messages", messageHandler.SendMessage)
		http.GET("/chats/:id/messages", messageHandler.GetMessages)
		http.PUT("/messages/:id", messageHandler.UpdateMessage)
		http.DELETE("/messages/:id", messageHandler.DeleteMessage)
		http.POST("/messages/:id/read", messageHandler.MarkAsRead)

		// Файлы
		http.POST("/messages/:id/files", fileHandler.UploadFile)
		http.GET("/files/:id", fileHandler.GetFile)
		http.DELETE("/files/:id", fileHandler.DeleteFile)

		// Поиск
		http.GET("/search/messages", searchHandler.SearchMessages)
		http.GET("/search/chats", searchHandler.SearchChats)
	}

	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
