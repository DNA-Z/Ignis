package handlers

import (
	"net/http"
	"strconv"

	"github.com/DNA-Z/Ignis/internal/service"

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

func (h *SearchHandler) SearchMessages(c *gin.Context) {
	userID, err := getUserID(c)
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

func (h *SearchHandler) SearchChats(c *gin.Context) {
	userID, err := getUserID(c)
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
