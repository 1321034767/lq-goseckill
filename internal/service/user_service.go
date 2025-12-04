package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/example/goseckill/internal/auth"
	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/user"
)

type UserService struct {
	repo user.Repository
	jwt  *config.JWTConfig
}

func NewUserService(repo user.Repository, jwt *config.JWTConfig) *UserService {
	return &UserService{repo: repo, jwt: jwt}
}

func hashPassword(raw, salt string) string {
	h := sha256.Sum256([]byte(raw + salt))
	return hex.EncodeToString(h[:])
}

// Register 简单注册（示例用）
func (s *UserService) Register(ctx context.Context, username, password string) (*user.User, error) {
	u := &user.User{
		Username: username,
		Salt:     "goseckill", // 简化实现，真实业务请使用随机盐
	}
	u.Password = hashPassword(password, u.Salt)
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login 登录并返回 JWT
func (s *UserService) Login(ctx context.Context, username, password string) (string, error) {
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if hashPassword(password, u.Salt) != u.Password {
		return "", errors.New("invalid password")
	}
	return auth.GenerateToken(s.jwt, u.ID, u.Username)
}
