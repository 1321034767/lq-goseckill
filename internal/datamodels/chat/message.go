package chat

import (
	"context"
	"time"
)

// Message 聊天消息模型（简单 demo，用于 Admin <-> 用户聊天）
type Message struct {
	ID        uint64    `gorm:"primaryKey"`
	ContactID string    `gorm:"size:64;index;not null"` // 会话标识，例如用户 ID 或用户名
	From      string    `gorm:"size:16;not null"`       // "self" 或 "friend"
	Content   string    `gorm:"size:512;not null"`
	CreatedAt time.Time `gorm:"index"`
}

// Repository 聊天消息仓储接口
type Repository interface {
	ListByContact(ctx context.Context, contactID string, afterID uint64, limit int) ([]*Message, error)
	Create(ctx context.Context, m *Message) error
}

