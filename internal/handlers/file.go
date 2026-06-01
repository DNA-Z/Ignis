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

// UploadFile загружает файл
// @Summary Загрузка файла
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID сообщения"
// @Param file formData file true "Файл для загрузки"
// @Success 201 {object} models.FileResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Router /api/messages/{id}/files [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
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

// GetFile скачивает файл
// @Summary Скачивание файла
// @Tags files
// @Produce octet-stream
// @Security BearerAuth
// @Param id path string true "ID файла"
// @Success 200 {file} binary
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id} [get]
func (h *FileHandler) GetFile(c *gin.Context) {
	userID, err := getAuthUserID(c)
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

// DeleteFile удаляет файл
// @Summary Удаление файла
// @Tags files
// @Security BearerAuth
// @Param id path string true "ID файла"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID, err := getAuthUserID(c)
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
