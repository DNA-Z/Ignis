package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ChatID    uuid.UUID  `json:"chat_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Text      string     `json:"text" gorm:"type:text"`
	IsDeleted bool       `json:"is_deleted" gorm:"default:false"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type SendMessageRequest struct {
	Text string `json:"text" binding:"required,min=1,max=5000"`
}

type UpdateMessageRequest struct {
	Text string `json:"text" binding:"required,min=1,max=5000"`
}

type MessageResponse struct {
	ID        uuid.UUID       `json:"id"`
	ChatID    uuid.UUID       `json:"chat_id"`
	User      *UserResponse   `json:"user"`
	Text      string          `json:"text"`
	IsDeleted bool            `json:"is_deleted"`
	CreatedAt time.Time       `json:"created_at"`
	EditedAt  *time.Time      `json:"edited_at,omitempty"`
	Files     []*FileResponse `json:"files,omitempty"`
}

type MessagesListResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

type ReadReceipt struct {
	MessageID uuid.UUID `json:"message_id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	ChatID    uuid.UUID `json:"chat_id" gorm:"type:uuid;index"`
	ReadAt    time.Time `json:"read_at" gorm:"autoCreateTime"`
}
