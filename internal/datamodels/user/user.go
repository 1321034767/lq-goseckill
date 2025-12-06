package user

import (
	"context"
	"time"
)

// User 用户模型
type User struct {
	ID        int64     `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;size:64;not null"`
	Password  string    `gorm:"size:255;not null"` // 已加密密码
	Salt      string    `gorm:"size:64"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Repository 用户仓储接口
type Repository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, u *User) error
	ListAll(ctx context.Context) ([]*User, error)
}


