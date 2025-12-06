package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/example/goseckill/internal/datamodels/account"
	"github.com/example/goseckill/internal/datamodels/order"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/datamodels/user"
	"github.com/example/goseckill/internal/repository/mysql"
)

// AccountService 提供账户余额与交易能力，并内置购买逻辑
type AccountService struct {
	db          *gorm.DB
	accountRepo account.Repository
	productRepo product.Repository
	orderRepo   order.Repository
	userRepo    user.Repository
}

// NewAccountService 创建账户服务
func NewAccountService(db *gorm.DB, productRepo product.Repository, orderRepo order.Repository, userRepo user.Repository) *AccountService {
	return &AccountService{
		db:          db,
		accountRepo: mysql.NewAccountRepository(db),
		productRepo: productRepo,
		orderRepo:   orderRepo,
		userRepo:    userRepo,
	}
}

// GetSummary 返回账户余额与冻结金额
func (s *AccountService) GetSummary(ctx context.Context, userID int64) (*account.Account, error) {
	acc, err := s.accountRepo.UpsertByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	// 为 admin 账户初始化 100 元基金（仅首次）
	if acc.Balance == 0 && acc.UserID == userID && s.userRepo != nil {
		u, err := s.userRepo.GetByID(ctx, userID)
		if err == nil && u != nil && u.Username == "admin" {
			acc.Balance = 10000 // 100 元
			if err := s.accountRepo.Update(ctx, acc); err != nil {
				return nil, err
			}
			_ = s.accountRepo.CreateTransaction(ctx, &account.Transaction{
				UserID: userID,
				Amount: 10000,
				Type:   "gift",
				Status: "success",
				Note:   "新用户赠送",
			})
		}
	}
	return acc, nil
}

// ListTransactions 查询交易流水
func (s *AccountService) ListTransactions(ctx context.Context, userID int64, limit int) ([]*account.Transaction, error) {
	return s.accountRepo.ListTransactions(ctx, userID, limit)
}

// AccountSummary 提供用户名+余额信息
type AccountSummary struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Balance  int64  `json:"balance"`
	Frozen   int64  `json:"frozen"`
	Updated  string `json:"updated_at"`
}

// ListAccounts 获取所有用户的余额信息（确保账户存在）
func (s *AccountService) ListAccounts(ctx context.Context) ([]AccountSummary, error) {
	users, err := s.userRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AccountSummary, 0, len(users))
	for _, u := range users {
		acc, err := s.accountRepo.UpsertByUserID(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, AccountSummary{
			UserID:   u.ID,
			Username: u.Username,
			Balance:  acc.Balance,
			Frozen:   acc.Frozen,
			Updated:  acc.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out, nil
}

// ListOrdersByUser 查询指定用户的订单
func (s *AccountService) ListOrdersByUser(ctx context.Context, userID int64) ([]*order.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID)
}

// Purchase 购买商品（扣余额、减库存、生成订单、写流水）
func (s *AccountService) Purchase(ctx context.Context, userID, productID, qty int64) (*order.Order, error) {
	if qty <= 0 {
		return nil, errors.New("数量必须大于 0")
	}

	var resultOrder *order.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 锁定/创建账户
		var acc account.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&acc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				acc = account.Account{UserID: userID}
				if err := tx.Create(&acc).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// 2) 锁定商品
		var p product.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&p, productID).Error; err != nil {
			return fmt.Errorf("商品不存在: %w", err)
		}
		if p.Status != 1 {
			return fmt.Errorf("商品不可购买")
		}
		if p.Stock < qty {
			return fmt.Errorf("库存不足")
		}

		// 3) 计算总价并校验余额
		total := p.Price * qty
		if acc.Balance < total {
			return fmt.Errorf("余额不足，需 ¥%.2f，当前 ¥%.2f", float64(total)/100, float64(acc.Balance)/100)
		}

		// 4) 扣减余额与库存
		acc.Balance -= total
		if err := tx.Save(&acc).Error; err != nil {
			return err
		}

		p.Stock -= qty
		if err := tx.Save(&p).Error; err != nil {
			return err
		}

		// 5) 创建订单
		o := order.Order{
			UserID:    userID,
			ProductID: productID,
			Price:     total,
			Status:    1, // 已支付
		}
		if err := tx.Create(&o).Error; err != nil {
			return err
		}
		resultOrder = &o

		// 6) 写交易流水
		if err := tx.Create(&account.Transaction{
			UserID: userID,
			Amount: -total,
			Type:   "purchase",
			Status: "success",
			Note:   fmt.Sprintf("订单 #%d", o.ID),
		}).Error; err != nil {
			return err
		}

		return nil
	})

	return resultOrder, err
}

// SeckillCharge 秒杀扣费（不再操作商品表，只扣减余额、创建订单和流水）
// price 单位为分，调用方需要自行根据秒杀折扣计算好价格。
func (s *AccountService) SeckillCharge(ctx context.Context, userID, productID, price int64) (*order.Order, error) {
	if price <= 0 {
		return nil, errors.New("价格必须大于 0")
	}

	var resultOrder *order.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 锁定/创建账户
		var acc account.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&acc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				acc = account.Account{UserID: userID}
				if err := tx.Create(&acc).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// 2) 校验余额
		if acc.Balance < price {
			return fmt.Errorf("余额不足，需 ¥%.2f，当前 ¥%.2f", float64(price)/100, float64(acc.Balance)/100)
		}

		// 3) 扣减余额
		acc.Balance -= price
		if err := tx.Save(&acc).Error; err != nil {
			return err
		}

		// 4) 创建订单（状态为已支付）
		o := order.Order{
			UserID:    userID,
			ProductID: productID,
			Price:     price,
			Status:    1, // 已支付
		}
		if err := tx.Create(&o).Error; err != nil {
			return err
		}
		resultOrder = &o

		// 5) 写交易流水
		if err := tx.Create(&account.Transaction{
			UserID: userID,
			Amount: -price,
			Type:   "seckill",
			Status: "success",
			Note:   fmt.Sprintf("秒杀订单 #%d", o.ID),
		}).Error; err != nil {
			return err
		}

		return nil
	})

	return resultOrder, err
}

// Recharge 简单充值示例，方便测试
func (s *AccountService) Recharge(ctx context.Context, userID, amount int64) (*account.Account, error) {
	if amount <= 0 {
		return nil, errors.New("充值金额需大于 0")
	}
	var acc account.Account
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&acc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				acc = account.Account{UserID: userID}
				if err := tx.Create(&acc).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		acc.Balance += amount
		if err := tx.Save(&acc).Error; err != nil {
			return err
		}
		if err := tx.Create(&account.Transaction{
			UserID: userID,
			Amount: amount,
			Type:   "recharge",
			Status: "success",
			Note:   "手动充值",
		}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &acc, nil
}
