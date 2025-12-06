package account

import (
	"context"
	"time"
)

// Account 用户账户余额
type Account struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"uniqueIndex;not null"`
	Balance   int64     `gorm:"not null"` // 可用余额，单位：分
	Frozen    int64     `gorm:"not null"` // 冻结金额，单位：分
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Transaction 账户交易流水
type Transaction struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"index;not null"`
	Amount    int64     `gorm:"not null"`       // 正数入账，负数出账，单位分
	Type      string    `gorm:"size:32;index"`  // purchase / refund / recharge 等
	Status    string    `gorm:"size:32;index"`  // success / failed / pending
	Note      string    `gorm:"size:255"`       // 备注
	CreatedAt time.Time `gorm:"index"`
}

// Repository 账户仓储接口
type Repository interface {
	GetByUserID(ctx context.Context, userID int64) (*Account, error)
	UpsertByUserID(ctx context.Context, userID int64) (*Account, error)
	Update(ctx context.Context, acc *Account) error

	CreateTransaction(ctx context.Context, tx *Transaction) error
	ListTransactions(ctx context.Context, userID int64, limit int) ([]*Transaction, error)
	ListAll(ctx context.Context) ([]*Account, error)
}
