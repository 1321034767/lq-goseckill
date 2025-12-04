package service

import (
	"context"

	"github.com/example/goseckill/internal/datamodels/chat"
)

// ChatService 聊天服务（封装基础的消息读写）
type ChatService struct {
	repo chat.Repository
}

// NewChatService 创建聊天服务
func NewChatService(repo chat.Repository) *ChatService {
	return &ChatService{repo: repo}
}

// ListMessages 返回某个会话的消息列表
func (s *ChatService) ListMessages(ctx context.Context, contactID string, afterID uint64, limit int) ([]*chat.Message, error) {
	return s.repo.ListByContact(ctx, contactID, afterID, limit)
}

// SendMessage 发送一条消息
func (s *ChatService) SendMessage(ctx context.Context, contactID, from, content string) (*chat.Message, error) {
	m := &chat.Message{
		ContactID: contactID,
		From:      from,
		Content:   content,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

