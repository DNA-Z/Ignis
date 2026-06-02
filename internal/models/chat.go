package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatType string

const (
	PrivateChat ChatType = "private"
	ChannelChat ChatType = "channel"
)

type Chat struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type        ChatType   `json:"type" gorm:"type:varchar(20);not null"`
	Name        string     `json:"name" gorm:"size:200"`
	Description string     `json:"description" gorm:"size:500"`
	CreatedBy   uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

type CreateChatRequest struct {
	Type        ChatType    `json:"type" binding:"required,oneof=private channel"`
	Name        string      `json:"name" binding:"required_if=Type channel,omitempty,min=1,max=200"`
	Description string      `json:"description" binding:"max=500"`
	UserIDs     []uuid.UUID `json:"user_ids" binding:"required,min=1"`
}

type UpdateChatRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

type ChatResponse struct {
	ID           uuid.UUID        `json:"id"`
	Type         ChatType         `json:"type"`
	Name         string           `json:"name,omitempty"`
	Description  string           `json:"description,omitempty"`
	CreatedBy    uuid.UUID        `json:"created_by"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Participants []UserResponse   `json:"participants,omitempty"`
	LastMessage  *MessageResponse `json:"last_message,omitempty"`
	UnreadCount  int64            `json:"unread_count,omitempty"`
}
