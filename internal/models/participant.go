package models

import (
	"github.com/google/uuid"
	"time"
)

type ParticipantRole string

const (
	RoleMember ParticipantRole = "member"
	RoleAdmin  ParticipantRole = "admin"
)

type ChatParticipant struct {
	ChatID     uuid.UUID       `json:"chat_id" gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID       `json:"user_id" gorm:"type:uuid;primaryKey"`
	Role       ParticipantRole `json:"role" gorm:"type:varchar(20);default:'member'"`
	JoinedAt   time.Time       `json:"joined_at" gorm:"autoCreateTime"`
	LastReadAt *time.Time      `json:"last_read_at,omitempty"`
}

type AddParticipantRequest struct {
	UserIDs []uuid.UUID `json:"user_ids" binding:"required,min=1"`
}

type UpdateRoleRequest struct {
	Role ParticipantRole `json:"role" binding:"required,oneof=member admin"`
}
