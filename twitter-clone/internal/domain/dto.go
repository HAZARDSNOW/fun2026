package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// Common errors
var (
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPhone       = errors.New("invalid phone format")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBlocked        = errors.New("user is blocked")
	ErrInsufficientTokens = errors.New("insufficient tokens")
	ErrPostNotFound       = errors.New("post not found")
	ErrCannotLikeOwnPost  = errors.New("cannot like your own post")
	ErrAlreadyFollowing   = errors.New("already following this user")
	ErrNotFollowing       = errors.New("not following this user")
	ErrCannotFollowSelf   = errors.New("cannot follow yourself")
	ErrPostLocked         = errors.New("post is locked")
	ErrPostDeleted        = errors.New("post is deleted")
	ErrMaxRepliesReached  = errors.New("maximum replies reached")
	ErrReportNotFound     = errors.New("report not found")
)

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	FullName  string    `json:"full_name"`
	BirthDate time.Time `json:"birth_date"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Badge    string `json:"badge"`
}

// CreatePostRequest represents a create post request
type CreatePostRequest struct {
	Text      string `json:"text"`
	MediaType string `json:"media_type"` // "image", "video", or ""
	MediaPath string `json:"media_path"`
}

// FollowRequest represents a follow/unfollow request
type FollowRequest struct {
	TargetUserID uint64 `json:"target_user_id"`
}

// LikeRequest represents a like/unlike request
type LikeRequest struct {
	PostID uint64 `json:"post_id"`
}

// UpgradeSubscriptionRequest represents a subscription upgrade request
type UpgradeSubscriptionRequest struct {
	Plan string `json:"plan"` // "blue" or "gold"
}

// AddTokensRequest represents a request to add tokens
type AddTokensRequest struct {
	Amount int64 `json:"amount"`
}

// ReportRequest represents a report request
type ReportRequest struct {
	ContentID      uint64 `json:"content_id"`
	ReportedUserID uint64 `json:"reported_user_id"`
	Description    string `json:"description"`
}

// ValidateRegisterRequest validates the registration request
func ValidateRegisterRequest(req RegisterRequest) error {
	// Validate username
	if req.Username == "" {
		return errors.New("username is required")
	}

	// Validate email using regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return ErrInvalidEmail
	}

	// Validate phone using regex
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	if !phoneRegex.MatchString(req.Phone) {
		return ErrInvalidPhone
	}

	// Validate password (at least 8 characters)
	if len(req.Password) < 8 {
		return ErrInvalidPassword
	}

	return nil
}

// ExtractHashtags extracts hashtags from text
func ExtractHashtags(text string) []string {
	hashtagRegex := regexp.MustCompile(`#\w+`)
	matches := hashtagRegex.FindAllString(text, -1)
	
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, strings.ToLower(match))
	}
	
	return result
}

// GenerateShareLink generates a share link for a post
func GenerateShareLink(postID uint64) string {
	return "https://x.com/Posts/" + string(rune(postID))
}

// IsWeakPassword checks if password is weak
func IsWeakPassword(password string) bool {
	// Weak password: less than 8 characters or no variety
	if len(password) < 8 {
		return true
	}
	
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	
	return !(hasLower && hasUpper && hasDigit)
}
