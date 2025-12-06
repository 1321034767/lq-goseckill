package mysql

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/example/goseckill/internal/datamodels/account"
)

type accountRepo struct {
	db *gorm.DB
}

// NewAccountRepository 创建账户仓储
func NewAccountRepository(db *gorm.DB) account.Repository {
	return &accountRepo{db: db}
}

func (r *accountRepo) GetByUserID(ctx context.Context, userID int64) (*account.Account, error) {
	var acc account.Account
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

// UpsertByUserID 确保账户存在，返回账户实例
func (r *accountRepo) UpsertByUserID(ctx context.Context, userID int64) (*account.Account, error) {
	acc := account.Account{UserID: userID}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoNothing: true,
	}).Create(&acc).Error; err != nil {
		return nil, err
	}
	// 再次查询
	return r.GetByUserID(ctx, userID)
}

func (r *accountRepo) Update(ctx context.Context, acc *account.Account) error {
	return r.db.WithContext(ctx).Save(acc).Error
}

func (r *accountRepo) CreateTransaction(ctx context.Context, tx *account.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *accountRepo) ListTransactions(ctx context.Context, userID int64, limit int) ([]*account.Transaction, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []*account.Transaction
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *accountRepo) ListAll(ctx context.Context) ([]*account.Account, error) {
	var list []*account.Account
	if err := r.db.WithContext(ctx).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
