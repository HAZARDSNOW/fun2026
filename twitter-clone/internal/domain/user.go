package domain

import (
	"time"
)

// Account represents the base account entity (cannot be instantiated directly)
type Account struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	FullName  string    `json:"full_name"`
	BirthDate time.Time `json:"birth_date"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CoverURL  string    `json:"cover_url"`
	JoinedAt  time.Time `json:"joined_at"`
}

// User represents the base user entity (cannot be instantiated directly)
type User struct {
	Account
	Credit       float64   `json:"credit"`
	Token        int64     `json:"token"`
	Bio          string    `json:"bio"`
	PostIDs      []uint64  `json:"post_ids"`
	FollowerIDs  []uint64  `json:"follower_ids"`
	FollowingIDs []uint64  `json:"following_ids"`
	LikedPostIDs []uint64  `json:"liked_post_ids"`
	Badge        string    `json:"badge"`
	Blocked      bool      `json:"blocked"`
	IsDeleted    bool      `json:"is_deleted"`
}

// NormalUser represents a regular user
type NormalUser struct {
	User
}

// PremiumUser represents a premium user (cannot be instantiated directly)
type PremiumUser struct {
	User
	SubscriptionEnd time.Time `json:"subscription_end"`
}

// BlueUser represents a blue premium user
type BlueUser struct {
	PremiumUser
}

// GoldUser represents a gold premium user
type GoldUser struct {
	PremiumUser
}

// GetUserBadge returns the badge of the user
func (u *User) GetUserBadge() string {
	return u.Badge
}

// IsPremium checks if user is premium
func (u *User) IsPremium() bool {
	return u.Badge == "blue" || u.Badge == "gold"
}

// NewNormalUser creates a new normal user
func NewNormalUser(account Account) *NormalUser {
	return &NormalUser{
		User: User{
			Account:        account,
			Credit:         0,
			Token:          0,
			Bio:            "",
			PostIDs:        []uint64{},
			FollowerIDs:    []uint64{},
			FollowingIDs:   []uint64{},
			LikedPostIDs:   []uint64{},
			Badge:          "",
			Blocked:        false,
			IsDeleted:      false,
		},
	}
}

// NewBlueUser creates a new blue user
func NewBlueUser(account Account) *BlueUser {
	return &BlueUser{
		PremiumUser: PremiumUser{
			User: User{
				Account:        account,
				Credit:         0,
				Token:          3000,
				Bio:            "",
				PostIDs:        []uint64{},
				FollowerIDs:    []uint64{},
				FollowingIDs:   []uint64{},
				LikedPostIDs:   []uint64{},
				Badge:          "blue",
				Blocked:        false,
				IsDeleted:      false,
			},
			SubscriptionEnd: time.Now().AddDate(0, 1, 0),
		},
	}
}

// NewGoldUser creates a new gold user
func NewGoldUser(account Account) *GoldUser {
	return &GoldUser{
		PremiumUser: PremiumUser{
			User: User{
				Account:        account,
				Credit:         0,
				Token:          3000,
				Bio:            "",
				PostIDs:        []uint64{},
				FollowerIDs:    []uint64{},
				FollowingIDs:   []uint64{},
				LikedPostIDs:   []uint64{},
				Badge:          "gold",
				Blocked:        false,
				IsDeleted:      false,
			},
			SubscriptionEnd: time.Now().AddDate(0, 1, 0),
		},
	}
}

// UpgradeToBlue upgrades user to blue subscription
func (n *NormalUser) UpgradeToBlue() *BlueUser {
	return &BlueUser{
		PremiumUser: PremiumUser{
			User: User{
				Account:        n.Account,
				Credit:         n.Credit,
				Token:          n.Token + 3000,
				Bio:            n.Bio,
				PostIDs:        n.PostIDs,
				FollowerIDs:    n.FollowerIDs,
				FollowingIDs:   n.FollowingIDs,
				LikedPostIDs:   n.LikedPostIDs,
				Badge:          "blue",
				Blocked:        n.Blocked,
				IsDeleted:      n.IsDeleted,
			},
			SubscriptionEnd: time.Now().AddDate(0, 1, 0),
		},
	}
}

// UpgradeToGold upgrades user to gold subscription
func (n *NormalUser) UpgradeToGold() *GoldUser {
	return &GoldUser{
		PremiumUser: PremiumUser{
			User: User{
				Account:        n.Account,
				Credit:         n.Credit,
				Token:          n.Token + 3000,
				Bio:            n.Bio,
				PostIDs:        n.PostIDs,
				FollowerIDs:    n.FollowerIDs,
				FollowingIDs:   n.FollowingIDs,
				LikedPostIDs:   n.LikedPostIDs,
				Badge:          "gold",
				Blocked:        n.Blocked,
				IsDeleted:      n.IsDeleted,
			},
			SubscriptionEnd: time.Now().AddDate(0, 1, 0),
		},
	}
}

// CalculatePostCost calculates the cost of creating a post based on user type
func (n *NormalUser) CalculatePostCost(textLen int, hasMedia bool) int64 {
	cost := int64(textLen)
	if hasMedia {
		cost += 10
	}
	return cost
}

func (b *BlueUser) CalculatePostCost(textLen int, hasMedia bool) int64 {
	cost := int64(textLen) / 2
	if hasMedia {
		cost += 5
	}
	return cost
}

func (g *GoldUser) CalculatePostCost(textLen int, hasMedia bool) int64 {
	return 5 // Fixed cost for gold users
}

// Admin represents the singleton admin user
type Admin struct {
	Account
}

var adminInstance *Admin

// GetAdminInstance returns the singleton admin instance
func GetAdminInstance() *Admin {
	if adminInstance == nil {
		adminInstance = &Admin{}
	}
	return adminInstance
}
