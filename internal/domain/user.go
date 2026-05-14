package domain

import (
	"time"
)

// Account represents the base account structure (Abstract class equivalent)
type Account struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	FullName  string    `json:"full_name"`
	BirthDate time.Time `json:"birth_date"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CoverURL  string    `json:"cover_url"`
	JoinDate  time.Time `json:"join_date"`
}

// User interface defines common user behaviors
type User interface {
	GetID() int64
	GetUsername() string
	GetBadge() string
	GetAccountType() string
	CalculatePostCost(textLen int, hasMedia bool) int
	CanEditPost() bool
	GetDailyTokenBonus() int
}

// NormalUser represents a regular user
type NormalUser struct {
	Account
	Balance       int     `json:"balance"`
	Token         int     `json:"token"`
	Bio           string  `json:"bio"`
	Posts         []int64 `json:"posts"`
	Followers     []int64 `json:"followers"`
	Following     []int64 `json:"following"`
	LikedPosts    []int64 `json:"liked_posts"`
	Badge         string  `json:"badge"`
	Blocked       bool    `json:"blocked"`
	InterestTags  []int64 `json:"interest_tags"`
}

func (u *NormalUser) GetID() int64                    { return u.ID }
func (u *NormalUser) GetUsername() string             { return u.Username }
func (u *NormalUser) GetBadge() string                { return u.Badge }
func (u *NormalUser) GetAccountType() string          { return "normal" }
func (u *NormalUser) CanEditPost() bool               { return false }
func (u *NormalUser) GetDailyTokenBonus() int         { return 250 }
func (u *NormalUser) GetBlocked() bool                { return u.Blocked }
func (u *NormalUser) CalculatePostCost(textLen int, hasMedia bool) int {
	cost := textLen
	if hasMedia {
		cost += 10
	}
	return cost
}

// PremiumUser is the base for premium users (Blue and Gold)
type PremiumUser struct {
	NormalUser
	SubscriptionEnd time.Time `json:"subscription_end"`
}

// BlueUser represents Blue subscription user
type BlueUser struct {
	PremiumUser
}

func (u *BlueUser) GetBadge() string        { return "blue" }
func (u *BlueUser) GetAccountType() string  { return "blue" }
func (u *BlueUser) CanEditPost() bool       { return true }
func (u *BlueUser) GetDailyTokenBonus() int { return 400 }
func (u *BlueUser) CalculatePostCost(textLen int, hasMedia bool) int {
	cost := textLen / 2
	if hasMedia {
		cost += 5
	}
	return cost
}

// GoldUser represents Gold subscription user
type GoldUser struct {
	PremiumUser
}

func (u *GoldUser) GetBadge() string        { return "gold" }
func (u *GoldUser) GetAccountType() string  { return "gold" }
func (u *GoldUser) CanEditPost() bool       { return true }
func (u *GoldUser) GetDailyTokenBonus() int { return 600 }
func (u *GoldUser) CalculatePostCost(textLen int, hasMedia bool) int {
	return 5 // Fixed cost for gold users
}

// Admin is a Singleton
type Admin struct {
	Account
}

var adminInstance *Admin

func GetAdminInstance() *Admin {
	if adminInstance == nil {
		adminInstance = &Admin{
			Account: Account{
				ID:       0,
				Username: "admin",
			},
		}
	}
	return adminInstance
}
