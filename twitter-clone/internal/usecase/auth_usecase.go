package usecase

import (
	"twitter-clone/internal/domain"
	"twitter-clone/internal/infrastructure/repository"
	"twitter-clone/pkg/hash"
	"twitter-clone/pkg/token"
)

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	userRepo repository.UserRepository
}

// NewAuthUseCase creates a new AuthUseCase
func NewAuthUseCase(userRepo repository.UserRepository) *AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
	}
}

// Register registers a new user
func (uc *AuthUseCase) Register(req domain.RegisterRequest) (*domain.User, error) {
	// Validate request
	if err := domain.ValidateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Check if username exists
	if uc.userRepo.ExistsByUsername(req.Username) {
		return nil, domain.ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create account
	account := domain.Account{
		Username:  req.Username,
		Password:  hashedPassword,
		FullName:  req.FullName,
		BirthDate: req.BirthDate,
		Email:     req.Email,
		Phone:     req.Phone,
		JoinedAt:  domain.NewNormalUser(domain.Account{}).JoinedAt,
	}

	// Save to repository
	if err := uc.userRepo.Create(account); err != nil {
		return nil, err
	}

	// Get the created user
	user, err := uc.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a token
func (uc *AuthUseCase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
	// Get user by username
	user, err := uc.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Check if user is blocked
	if user.Blocked {
		return nil, domain.ErrUserBlocked
	}

	// Verify password
	if err := hash.VerifyPassword(user.Password, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate JWT token
	jwtToken, err := token.GenerateToken(user.ID, user.Username, user.Badge)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		UserID:   user.ID,
		Username: user.Username,
		Token:    jwtToken,
		Badge:    user.Badge,
	}, nil
}

// GetUserByID gets a user by ID
func (uc *AuthUseCase) GetUserByID(id uint64) (*domain.User, error) {
	return uc.userRepo.GetByID(id)
}

// GetUserByUsername gets a user by username
func (uc *AuthUseCase) GetUserByUsername(username string) (*domain.User, error) {
	return uc.userRepo.GetByUsername(username)
}

// UpdateUser updates a user
func (uc *AuthUseCase) UpdateUser(user *domain.User) error {
	return uc.userRepo.Update(user)
}

// SearchUsers searches users by username
func (uc *AuthUseCase) SearchUsers(query string) []*domain.User {
	return uc.userRepo.SearchByUsername(query)
}

// FollowUser makes a user follow another user
func (uc *AuthUseCase) FollowUser(followerID, targetID uint64) error {
	if followerID == targetID {
		return domain.ErrCannotFollowSelf
	}

	follower, err := uc.userRepo.GetByID(followerID)
	if err != nil {
		return err
	}

	target, err := uc.userRepo.GetByID(targetID)
	if err != nil {
		return err
	}

	// Check if already following
	for _, id := range follower.FollowingIDs {
		if id == targetID {
			return domain.ErrAlreadyFollowing
		}
	}

	// Add to following list
	follower.FollowingIDs = append(follower.FollowingIDs, targetID)
	
	// Add to followers list
	target.FollowerIDs = append(target.FollowerIDs, followerID)

	// Update both users
	if err := uc.userRepo.Update(follower); err != nil {
		return err
	}

	if err := uc.userRepo.Update(target); err != nil {
		return err
	}

	return nil
}

// UnfollowUser makes a user unfollow another user
func (uc *AuthUseCase) UnfollowUser(followerID, targetID uint64) error {
	follower, err := uc.userRepo.GetByID(followerID)
	if err != nil {
		return err
	}

	target, err := uc.userRepo.GetByID(targetID)
	if err != nil {
		return err
	}

	// Remove from following list
	found := false
	for i, id := range follower.FollowingIDs {
		if id == targetID {
			follower.FollowingIDs = append(follower.FollowingIDs[:i], follower.FollowingIDs[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return domain.ErrNotFollowing
	}

	// Remove from followers list
	for i, id := range target.FollowerIDs {
		if id == followerID {
			target.FollowerIDs = append(target.FollowerIDs[:i], target.FollowerIDs[i+1:]...)
			break
		}
	}

	// Update both users
	if err := uc.userRepo.Update(follower); err != nil {
		return err
	}

	if err := uc.userRepo.Update(target); err != nil {
		return err
	}

	return nil
}

// UpgradeSubscription upgrades a user's subscription
func (uc *AuthUseCase) UpgradeSubscription(userID uint64, plan string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user.IsPremium() {
		// Already premium, could implement upgrade logic here
		return nil
	}

	// For simplicity, we're using NormalUser type assertion
	// In real implementation, you'd handle this differently
	normalUser := &domain.NormalUser{User: *user}

	var newUser *domain.User
	if plan == "blue" {
		blueUser := normalUser.UpgradeToBlue()
		newUser = &blueUser.User
	} else if plan == "gold" {
		goldUser := normalUser.UpgradeToGold()
		newUser = &goldUser.User
	} else {
		return domain.ErrInvalidCredentials
	}

	return uc.userRepo.Update(newUser)
}

// AddTokens adds tokens to a user's account
func (uc *AuthUseCase) AddTokens(userID uint64, amount int64) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.Token += amount
	return uc.userRepo.Update(user)
}

// BlockUser blocks a user (admin function)
func (uc *AuthUseCase) BlockUser(adminID, userID uint64) error {
	// Verify admin (in real implementation, check admin rights)
	admin := domain.GetAdminInstance()
	if admin.ID != adminID {
		return domain.ErrInvalidCredentials
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.Blocked = true
	return uc.userRepo.Update(user)
}

// UnblockUser unblocks a user (admin function)
func (uc *AuthUseCase) UnblockUser(adminID, userID uint64) error {
	// Verify admin
	admin := domain.GetAdminInstance()
	if admin.ID != adminID {
		return domain.ErrInvalidCredentials
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.Blocked = false
	return uc.userRepo.Update(user)
}
