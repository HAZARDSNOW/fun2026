package usecase

import (
	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/repository"
	"auth-service/pkg/hash"
	"auth-service/pkg/token"
	"errors"
	"time"
)

type AuthUsecase struct {
	userRepo   repository.UserRepository
	hasher     hash.PasswordHasher
	jwtManager *token.JWTManager
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	hasher hash.PasswordHasher,
	jwtManager *token.JWTManager,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		hasher:     hasher,
		jwtManager: jwtManager,
	}
}

func (uc *AuthUsecase) Register(input domain.RegisterInput) (*domain.AuthResponse, error) {
	if input.Email == "" || input.Password == "" {
		return nil, errors.New("email and password are required")
	}

	if len(input.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	hashedPassword, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		Email:     input.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, err
	}

	tokenString, err := uc.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &domain.AuthResponse{
		Token: tokenString,
		User:  *user,
	}, nil
}

func (uc *AuthUsecase) Login(input domain.LoginInput) (*domain.AuthResponse, error) {
	if input.Email == "" || input.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := uc.userRepo.GetByEmail(input.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !uc.hasher.Verify(input.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	tokenString, err := uc.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &domain.AuthResponse{
		Token: tokenString,
		User:  *user,
	}, nil
}
