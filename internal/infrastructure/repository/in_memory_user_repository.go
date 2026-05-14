package repository

import (
	"auth-service/internal/domain"
	"errors"
	"sync"
)

type InMemoryUserRepository struct {
	users  map[int64]*domain.User
	emailMap map[string]int64
	mu     sync.RWMutex
	nextID int64
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:    make(map[int64]*domain.User),
		emailMap: make(map[string]int64),
		nextID:   1,
	}
}

func (r *InMemoryUserRepository) Create(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emailMap[user.Email]; exists {
		return errors.New("user with this email already exists")
	}

	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = user
	r.emailMap[user.Email] = user.ID
	return nil
}

func (r *InMemoryUserRepository) GetByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.emailMap[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return r.users[id], nil
}

func (r *InMemoryUserRepository) GetByID(id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}
