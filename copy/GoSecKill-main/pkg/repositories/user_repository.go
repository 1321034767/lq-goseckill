package repositories

import (
	"GoSecKill/pkg/models"

	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUserList() ([]models.User, error)
	GetUserById(id int) (models.User, error)
	GetUserListByUsername(username string) ([]models.User, error)
	InsertUser(user models.User) error
	UpdateUser(user models.User) error
	DeleteUser(id int) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{db: db}
}

func (u UserRepository) GetUserList() ([]models.User, error) {
	var userList []models.User
	err := u.db.Find(&userList).Error
	if err != nil {
		return nil, err
	}
	return userList, nil
}

func (u UserRepository) GetUserById(id int) (models.User, error) {
	var userItem models.User
	err := u.db.First(&userItem, id).Error
	if err != nil {
		return models.User{}, err
	}
	return userItem, nil
}

func (u UserRepository) GetUserListByUsername(username string) ([]models.User, error) {
	var userList []models.User
	err := u.db.Where("username like ?", "%"+username+"%").Find(&userList).Error
	if err != nil {
		return nil, err
	}
	return userList, nil
}

func (u UserRepository) InsertUser(user models.User) error {
	err := u.db.Create(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (u UserRepository) UpdateUser(user models.User) error {
	err := u.db.Save(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (u UserRepository) DeleteUser(id int) error {
	err := u.db.Delete(&models.User{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
