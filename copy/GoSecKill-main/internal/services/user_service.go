package services

import (
	"GoSecKill/pkg/models"
	"GoSecKill/pkg/repositories"
)

type IUserService interface {
	GetUserList() ([]models.User, error)
	GetUserById(id int) (models.User, error)
	GetUserListByUsername(username string) ([]models.User, error)
	InsertUser(user models.User) error
	UpdateUser(user models.User) error
	DeleteUser(id int) error
}

type UserService struct {
	userRepository repositories.IUserRepository
}

func NewUserService(userRepository repositories.IUserRepository) IUserService {
	return &UserService{userRepository: userRepository}
}

func (u UserService) GetUserList() (users []models.User, err error) {
	return u.userRepository.GetUserList()
}

func (u UserService) GetUserById(id int) (user models.User, err error) {
	return u.userRepository.GetUserById(id)
}

func (u UserService) GetUserListByUsername(username string) (users []models.User, err error) {
	return u.userRepository.GetUserListByUsername(username)
}

func (u UserService) InsertUser(user models.User) (err error) {
	return u.userRepository.InsertUser(user)
}

func (u UserService) UpdateUser(user models.User) (err error) {
	return u.userRepository.UpdateUser(user)
}

func (u UserService) DeleteUser(id int) (err error) {
	return u.userRepository.DeleteUser(id)
}
