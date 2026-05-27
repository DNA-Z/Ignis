package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MessageID uuid.UUID `json:"message_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Name      string    `json:"name" gorm:"not null"`
	Size      int64     `json:"size" gorm:"not null"`
	MimeType  string    `json:"mime_type" gorm:"not null"`
	Path      string    `json:"path" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type FileResponse struct {
	ID        uuid.UUID `json:"file_id"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
