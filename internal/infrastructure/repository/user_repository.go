package repository

import (
	"auth-service/internal/domain"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByEmail(email string) (*domain.User, error)
	GetByID(id int64) (*domain.User, error)
}
