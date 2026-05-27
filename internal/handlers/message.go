package handlers

import (
	"github.com/DNA-Z/Ignis/internal/models"
	"github.com/DNA-Z/Ignis/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	messageService service.MessageService
}

func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// SendMessage отправляет сообщение
// @Summary Отправка сообщения
// @Tags messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID чата"
// @Param request body models.SendMessageRequest true "Текст сообщения"
// @Success 201 {object} models.MessageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/chats/{id}/messages [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
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

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.messageService.SendMessage(c.Request.Context(), userID, chatID, &req)
	if err != nil {
		if err == service.ErrNotParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessages получает сообщения чата
// @Summary Получение сообщений чата
// @Tags messages
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID чата"
// @Param limit query int false "Лимит" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} models.MessagesListResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/chats/{id}/messages [get]
func (h *MessageHandler) GetMessages(c *gin.Context) {
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

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	messages, err := h.messageService.GetMessages(c.Request.Context(), userID, chatID, limit, offset)
	if err != nil {
		if err == service.ErrNotParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// UpdateMessage обновляет сообщение
// @Summary Редактирование сообщения
// @Tags messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID сообщения"
// @Param request body models.UpdateMessageRequest true "Новый текст"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/messages/{id} [put]
func (h *MessageHandler) UpdateMessage(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	var req models.UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.messageService.UpdateMessage(c.Request.Context(), userID, messageID, &req); err != nil {
		if err == service.ErrNotMessageAuthor {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrMessageNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrMessageDeleted {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message updated successfully"})
}

// DeleteMessage удаляет сообщение
// @Summary Удаление сообщения
// @Tags messages
// @Security BearerAuth
// @Param id path string true "ID сообщения"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	if err := h.messageService.DeleteMessage(c.Request.Context(), userID, messageID); err != nil {
		if err == service.ErrNotMessageAuthor {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrMessageNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkAsRead отмечает сообщение как прочитанное
// @Summary Отметка о прочтении
// @Tags messages
// @Security BearerAuth
// @Param id path string true "ID сообщения"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/messages/{id}/read [post]
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	if err := h.messageService.MarkAsRead(c.Request.Context(), userID, messageID); err != nil {
		if err == service.ErrNotParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrMessageNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message marked as read"})
}
