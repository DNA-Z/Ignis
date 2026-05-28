package handlers

import (
	"github.com/DNA-Z/Ignis/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SearchHandler struct {
	searchService service.SearchService
}

func NewSearchHandler(searchService service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// SearchMessages ищет сообщения
// @Summary Поиск сообщений
// @Tags search
// @Produce json
// @Security BearerAuth
// @Param query string true "Поисковый запрос"
// @Param chatId query string false "ID чата (опционально)"
// @Param limit query int false "Лимит" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/search/messages [get]
func (h *SearchHandler) SearchMessages(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	var chatID *uuid.UUID
	if chatIDStr := c.Query("chatId"); chatIDStr != "" {
		id, err := uuid.Parse(chatIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
			return
		}
		chatID = &id
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	messages, total, err := h.searchService.SearchMessages(c.Request.Context(), userID, query, chatID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// SearchChats ищет чаты
// @Summary Поиск чатов
// @Tags search
// @Produce json
// @Security BearerAuth
// @Param query query string true "Поисковый запрос"
// @Param limit query int false "Лимит" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/search/chats [get]
func (h *SearchHandler) SearchChats(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	chats, total, err := h.searchService.SearchChats(c.Request.Context(), userID, query, limit, offset)
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
