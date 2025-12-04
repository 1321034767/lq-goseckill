package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/example/goseckill/internal/datamodels/chat"
)

type chatRepo struct {
	db *gorm.DB
}

// NewChatRepository 创建聊天消息仓储
func NewChatRepository(db *gorm.DB) chat.Repository {
	return &chatRepo{db: db}
}

func (r *chatRepo) ListByContact(ctx context.Context, contactID string, afterID uint64, limit int) ([]*chat.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var list []*chat.Message
	q := r.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Order("id ASC").
		Limit(limit)
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *chatRepo) Create(ctx context.Context, m *chat.Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

