package handlers

import (
	"net/http"
	"strconv"

	"github.com/DNA-Z/Ignis/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	fileService service.FileService
	maxFileSize int64
}

func NewFileHandler(fileService service.FileService, maxFileSize int64) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		maxFileSize: maxFileSize,
	}
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	if file.Size > h.maxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "file too large",
		})
		return
	}

	fileResponse, err := h.fileService.UploadFile(c.Request.Context(), userID, messageID, file)
	if err != nil {
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

	c.JSON(http.StatusCreated, fileResponse)
}

func (h *FileHandler) GetFile(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	file, reader, err := h.fileService.GetFile(c.Request.Context(), userID, fileID)
	if err != nil {
		if err == service.ErrNotParticipant {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+file.Name)
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))

	c.DataFromReader(http.StatusOK, file.Size, file.MimeType, reader, nil)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	if err := h.fileService.DeleteFile(c.Request.Context(), userID, fileID); err != nil {
		if err == service.ErrNotMessageAuthor {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
