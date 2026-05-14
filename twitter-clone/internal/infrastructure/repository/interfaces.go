package repository

import (
	"twitter-clone/internal/domain"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user domain.Account) error
	GetByID(id uint64) (*domain.User, error)
	GetByUsername(username string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id uint64) error
	ExistsByUsername(username string) bool
	GetAll() []*domain.User
	SearchByUsername(query string) []*domain.User
}

// PostRepository defines the interface for post data access
type PostRepository interface {
	Create(post *domain.Post) error
	GetByID(id uint64) (*domain.Post, error)
	GetByAuthorID(authorID uint64) []*domain.Post
	GetAll() []*domain.Post
	Update(post *domain.Post) error
	Delete(id uint64) error
	GetPopular(limit int) []*domain.Post
	SearchByText(query string) []*domain.Post
}

// HashtagRepository defines the interface for hashtag data access
type HashtagRepository interface {
	Create(hashtag *domain.Hashtag) error
	GetByID(id uint64) (*domain.Hashtag, error)
	GetByTitle(title string) (*domain.Hashtag, error)
	GetAll() []*domain.Hashtag
	GetPopular(limit int) []*domain.Hashtag
	Update(hashtag *domain.Hashtag) error
	AddPostToHashtag(hashtagID, postID uint64) error
}

// ReportRepository defines the interface for report data access
type ReportRepository interface {
	Create(report *domain.Report) error
	GetByID(id uint64) (*domain.Report, error)
	GetAll() []*domain.Report
	GetByStatus(status domain.ReportStatus) []*domain.Report
	Update(report *domain.Report) error
}

// Database represents the singleton database
type Database interface {
	UserRepository() UserRepository
	PostRepository() PostRepository
	HashtagRepository() HashtagRepository
	ReportRepository() ReportRepository
	Close() error
}
