package handlers

import (
	"github.com/DNA-Z/Ignis/internal/models"
	"github.com/DNA-Z/Ignis/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// CreateChat создаёт новый чат
// @Summary Создание чата
// @Tags chats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateChatRequest true "Данные для создания чата"
// @Success 201 {object} models.ChatResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/chats [post]
func (h *ChatHandler) CreateChat(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req models.CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, err := h.chatService.CreateChat(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

// GetChat возвращает информацию о чате
// @Summary Получение чата
// @Tags chats
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID чата"
// @Success 200 {object} models.ChatResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/chats/{id} [get]
func (h *ChatHandler) GetChat(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	chat, err := h.chatService.GetChat(c.Request.Context(), userID, chatID)
	if err != nil {
		if err == service.ErrNotParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chat)
}

// GetUserChats возвращает список чатов пользователя
// @Summary Список чатов пользователя
// @Tags chats
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Лимит" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {array} models.ChatResponse
// @Failure 401 {object} map[string]string
// @Router /api/chats [get]
func (h *ChatHandler) GetUserChats(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	chats, total, err := h.chatService.GetUserChats(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chats":  chats,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateChat обновляет чат
// @Summary Обновление чата
// @Tags chats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID чата"
// @Param request body models.UpdateChatRequest true "Данные для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/chats/{id} [put]
func (h *ChatHandler) UpdateChat(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	var req models.UpdateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chatService.UpdateChat(c.Request.Context(), userID, chatID, &req); err != nil {
		if err == service.ErrNotAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "chat updated successfully"})
}

// DeleteChat удаляет чат
// @Summary Удаление чата
// @Tags chats
// @Security BearerAuth
// @Param id path string true "ID чата"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/chats/{id} [delete]
func (h *ChatHandler) DeleteChat(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	if err := h.chatService.DeleteChat(c.Request.Context(), userID, chatID); err != nil {
		if err == service.ErrNotAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddParticipants добавляет участников в чат
// @Summary Добавление участников
// @Tags participants
// @Accept json
// @Security BearerAuth
// @Param id path string true "ID чата"
// @Param request body models.AddParticipantRequest true "ID пользователей"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/chats/{id}/users [post]
func (h *ChatHandler) AddParticipants(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	var req models.AddParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chatService.AddParticipants(c.Request.Context(), userID, chatID, &req); err != nil {
		if err == service.ErrNotAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "participants added successfully"})
}

// RemoveParticipant удаляет участника из чата
// @Summary Удаление участника
// @Tags participants
// @Security BearerAuth
// @Param id path string true "ID чата"
// @Param userId path string true "ID пользователя"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/chats/{id}/users/{userId} [delete]
func (h *ChatHandler) RemoveParticipant(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.chatService.RemoveParticipant(c.Request.Context(), userID, chatID, targetUserID); err != nil {
		if err == service.ErrNotAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "participant removed successfully"})
}

// UpdateRole обновляет роль участника
// @Summary Обновление роли
// @Tags participants
// @Accept json
// @Security BearerAuth
// @Param id path string true "ID чата"
// @Param userId path string true "ID пользователя"
// @Param request body models.UpdateRoleRequest true "Новая роль"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/chats/{id}/users/{userId}/role [put]
func (h *ChatHandler) UpdateRole(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	chatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chatService.UpdateRole(c.Request.Context(), userID, chatID, targetUserID, &req); err != nil {
		if err == service.ErrNotAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrChatNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}

func getAuthUserID(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, service.ErrInvalidCredentials
	}
	return userID.(uuid.UUID), nil
}
