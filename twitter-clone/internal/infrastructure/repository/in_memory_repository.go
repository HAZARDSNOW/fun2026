package repository

import (
	"strings"
	"sync"
	"twitter-clone/internal/domain"
)

// InMemoryDatabase implements the Database interface with in-memory storage
type InMemoryDatabase struct {
	users      map[uint64]*domain.User
	usernames  map[string]uint64
	posts      map[uint64]*domain.Post
	hashtags   map[uint64]*domain.Hashtag
	reports    map[uint64]*domain.Report
	userMutex  sync.RWMutex
	postMutex  sync.RWMutex
	tagMutex   sync.RWMutex
	reportMutex sync.RWMutex
	nextUserID uint64
	nextPostID uint64
	nextTagID  uint64
	nextReportID uint64
}

// NewInMemoryDatabase creates a new in-memory database
func NewInMemoryDatabase() *InMemoryDatabase {
	db := &InMemoryDatabase{
		users:      make(map[uint64]*domain.User),
		usernames:  make(map[string]uint64),
		posts:      make(map[uint64]*domain.Post),
		hashtags:   make(map[uint64]*domain.Hashtag),
		reports:    make(map[uint64]*domain.Report),
		nextUserID: 1,
		nextPostID: 1,
		nextTagID:  1,
		nextReportID: 1,
	}
	
	// Initialize some popular hashtags
	popularTags := []string{"#technology", "#news", "#sports", "#entertainment", "#music"}
	for _, tag := range popularTags {
		hashtag := domain.NewHashtag(db.nextTagID, tag)
		db.hashtags[db.nextTagID] = hashtag
		db.nextTagID++
	}
	
	return db
}

// UserRepository returns the user repository
func (db *InMemoryDatabase) UserRepository() UserRepository {
	return &InMemoryUserRepository{db: db}
}

// PostRepository returns the post repository
func (db *InMemoryDatabase) PostRepository() PostRepository {
	return &InMemoryPostRepository{db: db}
}

// HashtagRepository returns the hashtag repository
func (db *InMemoryDatabase) HashtagRepository() HashtagRepository {
	return &InMemoryHashtagRepository{db: db}
}

// ReportRepository returns the report repository
func (db *InMemoryDatabase) ReportRepository() ReportRepository {
	return &InMemoryReportRepository{db: db}
}

// Close closes the database connection (no-op for in-memory)
func (db *InMemoryDatabase) Close() error {
	return nil
}

// InMemoryUserRepository implements UserRepository with in-memory storage
type InMemoryUserRepository struct {
	db *InMemoryDatabase
}

// Create creates a new user
func (r *InMemoryUserRepository) Create(account domain.Account) error {
	r.db.userMutex.Lock()
	defer r.db.userMutex.Unlock()
	
	if _, exists := r.db.usernames[account.Username]; exists {
		return domain.ErrUsernameExists
	}
	
	user := domain.NewNormalUser(account)
	user.ID = r.db.nextUserID
	r.db.nextUserID++
	
	r.db.users[user.ID] = &user.User
	r.db.usernames[account.Username] = user.ID
	
	return nil
}

// GetByID gets a user by ID
func (r *InMemoryUserRepository) GetByID(id uint64) (*domain.User, error) {
	r.db.userMutex.RLock()
	defer r.db.userMutex.RUnlock()
	
	user, exists := r.db.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	
	return user, nil
}

// GetByUsername gets a user by username
func (r *InMemoryUserRepository) GetByUsername(username string) (*domain.User, error) {
	r.db.userMutex.RLock()
	defer r.db.userMutex.RUnlock()
	
	userID, exists := r.db.usernames[username]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	
	user, exists := r.db.users[userID]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	
	return user, nil
}

// Update updates a user
func (r *InMemoryUserRepository) Update(user *domain.User) error {
	r.db.userMutex.Lock()
	defer r.db.userMutex.Unlock()
	
	if _, exists := r.db.users[user.ID]; !exists {
		return domain.ErrUserNotFound
	}
	
	r.db.users[user.ID] = user
	return nil
}

// Delete deletes a user
func (r *InMemoryUserRepository) Delete(id uint64) error {
	r.db.userMutex.Lock()
	defer r.db.userMutex.Unlock()
	
	user, exists := r.db.users[id]
	if !exists {
		return domain.ErrUserNotFound
	}
	
	delete(r.db.usernames, user.Username)
	delete(r.db.users, id)
	
	return nil
}

// ExistsByUsername checks if a username exists
func (r *InMemoryUserRepository) ExistsByUsername(username string) bool {
	r.db.userMutex.RLock()
	defer r.db.userMutex.RUnlock()
	
	_, exists := r.db.usernames[username]
	return exists
}

// GetAll gets all users
func (r *InMemoryUserRepository) GetAll() []*domain.User {
	r.db.userMutex.RLock()
	defer r.db.userMutex.RUnlock()
	
	users := make([]*domain.User, 0, len(r.db.users))
	for _, user := range r.db.users {
		users = append(users, user)
	}
	
	return users
}

// SearchByUsername searches users by username
func (r *InMemoryUserRepository) SearchByUsername(query string) []*domain.User {
	r.db.userMutex.RLock()
	defer r.db.userMutex.RUnlock()
	
	query = strings.ToLower(query)
	var results []*domain.User
	
	for _, user := range r.db.users {
		if strings.Contains(strings.ToLower(user.Username), query) {
			results = append(results, user)
		}
	}
	
	return results
}

// InMemoryPostRepository implements PostRepository with in-memory storage
type InMemoryPostRepository struct {
	db *InMemoryDatabase
}

// Create creates a new post
func (r *InMemoryPostRepository) Create(post *domain.Post) error {
	r.db.postMutex.Lock()
	defer r.db.postMutex.Unlock()
	
	post.ID = r.db.nextPostID
	r.db.nextPostID++
	
	r.db.posts[post.ID] = post
	return nil
}

// GetByID gets a post by ID
func (r *InMemoryPostRepository) GetByID(id uint64) (*domain.Post, error) {
	r.db.postMutex.RLock()
	defer r.db.postMutex.RUnlock()
	
	post, exists := r.db.posts[id]
	if !exists {
		return nil, domain.ErrPostNotFound
	}
	
	return post, nil
}

// GetByAuthorID gets posts by author ID
func (r *InMemoryPostRepository) GetByAuthorID(authorID uint64) []*domain.Post {
	r.db.postMutex.RLock()
	defer r.db.postMutex.RUnlock()
	
	var posts []*domain.Post
	for _, post := range r.db.posts {
		if post.AuthorID == authorID && !post.Deleted {
			posts = append(posts, post)
		}
	}
	
	return posts
}

// GetAll gets all posts
func (r *InMemoryPostRepository) GetAll() []*domain.Post {
	r.db.postMutex.RLock()
	defer r.db.postMutex.RUnlock()
	
	posts := make([]*domain.Post, 0, len(r.db.posts))
	for _, post := range r.db.posts {
		if !post.Deleted {
			posts = append(posts, post)
		}
	}
	
	return posts
}

// Update updates a post
func (r *InMemoryPostRepository) Update(post *domain.Post) error {
	r.db.postMutex.Lock()
	defer r.db.postMutex.Unlock()
	
	if _, exists := r.db.posts[post.ID]; !exists {
		return domain.ErrPostNotFound
	}
	
	r.db.posts[post.ID] = post
	return nil
}

// Delete deletes a post
func (r *InMemoryPostRepository) Delete(id uint64) error {
	r.db.postMutex.Lock()
	defer r.db.postMutex.Unlock()
	
	post, exists := r.db.posts[id]
	if !exists {
		return domain.ErrPostNotFound
	}
	
	post.MarkAsDeleted()
	return nil
}

// GetPopular gets popular posts sorted by likes
func (r *InMemoryPostRepository) GetPopular(limit int) []*domain.Post {
	r.db.postMutex.RLock()
	defer r.db.postMutex.RUnlock()
	
	posts := make([]*domain.Post, 0)
	for _, post := range r.db.posts {
		if !post.Deleted && !post.Locked {
			posts = append(posts, post)
		}
	}
	
	// Sort by like count (simple bubble sort for demo)
	for i := 0; i < len(posts)-1; i++ {
		for j := i + 1; j < len(posts); j++ {
			if posts[j].LikeCount > posts[i].LikeCount {
				posts[i], posts[j] = posts[j], posts[i]
			}
		}
	}
	
	if limit > len(posts) {
		limit = len(posts)
	}
	
	return posts[:limit]
}

// SearchByText searches posts by text content
func (r *InMemoryPostRepository) SearchByText(query string) []*domain.Post {
	r.db.postMutex.RLock()
	defer r.db.postMutex.RUnlock()
	
	query = strings.ToLower(query)
	var results []*domain.Post
	
	for _, post := range r.db.posts {
		if !post.Deleted && strings.Contains(strings.ToLower(post.Text), query) {
			results = append(results, post)
		}
	}
	
	return results
}

// InMemoryHashtagRepository implements HashtagRepository with in-memory storage
type InMemoryHashtagRepository struct {
	db *InMemoryDatabase
}

// Create creates a new hashtag
func (r *InMemoryHashtagRepository) Create(hashtag *domain.Hashtag) error {
	r.db.tagMutex.Lock()
	defer r.db.tagMutex.Unlock()
	
	hashtag.ID = r.db.nextTagID
	r.db.nextTagID++
	
	r.db.hashtags[hashtag.ID] = hashtag
	return nil
}

// GetByID gets a hashtag by ID
func (r *InMemoryHashtagRepository) GetByID(id uint64) (*domain.Hashtag, error) {
	r.db.tagMutex.RLock()
	defer r.db.tagMutex.RUnlock()
	
	hashtag, exists := r.db.hashtags[id]
	if !exists {
		return nil, domain.ErrPostNotFound // Reusing error for simplicity
	}
	
	return hashtag, nil
}

// GetByTitle gets a hashtag by title
func (r *InMemoryHashtagRepository) GetByTitle(title string) (*domain.Hashtag, error) {
	r.db.tagMutex.RLock()
	defer r.db.tagMutex.RUnlock()
	
	title = strings.ToLower(title)
	for _, hashtag := range r.db.hashtags {
		if strings.ToLower(hashtag.Title) == title {
			return hashtag, nil
		}
	}
	
	return nil, domain.ErrPostNotFound
}

// GetAll gets all hashtags
func (r *InMemoryHashtagRepository) GetAll() []*domain.Hashtag {
	r.db.tagMutex.RLock()
	defer r.db.tagMutex.RUnlock()
	
	tags := make([]*domain.Hashtag, 0, len(r.db.hashtags))
	for _, tag := range r.db.hashtags {
		tags = append(tags, tag)
	}
	
	return tags
}

// GetPopular gets popular hashtags sorted by post count
func (r *InMemoryHashtagRepository) GetPopular(limit int) []*domain.Hashtag {
	r.db.tagMutex.RLock()
	defer r.db.tagMutex.RUnlock()
	
	tags := make([]*domain.Hashtag, 0, len(r.db.hashtags))
	for _, tag := range r.db.hashtags {
		tags = append(tags, tag)
	}
	
	// Sort by post count
	for i := 0; i < len(tags)-1; i++ {
		for j := i + 1; j < len(tags); j++ {
			if len(tags[j].PostIDs) > len(tags[i].PostIDs) {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}
	
	if limit > len(tags) {
		limit = len(tags)
	}
	
	return tags[:limit]
}

// Update updates a hashtag
func (r *InMemoryHashtagRepository) Update(hashtag *domain.Hashtag) error {
	r.db.tagMutex.Lock()
	defer r.db.tagMutex.Unlock()
	
	if _, exists := r.db.hashtags[hashtag.ID]; !exists {
		return domain.ErrPostNotFound
	}
	
	r.db.hashtags[hashtag.ID] = hashtag
	return nil
}

// AddPostToHashtag adds a post to a hashtag
func (r *InMemoryHashtagRepository) AddPostToHashtag(hashtagID, postID uint64) error {
	r.db.tagMutex.Lock()
	defer r.db.tagMutex.Unlock()
	
	hashtag, exists := r.db.hashtags[hashtagID]
	if !exists {
		return domain.ErrPostNotFound
	}
	
	hashtag.AddPost(postID)
	return nil
}

// InMemoryReportRepository implements ReportRepository with in-memory storage
type InMemoryReportRepository struct {
	db *InMemoryDatabase
}

// Create creates a new report
func (r *InMemoryReportRepository) Create(report *domain.Report) error {
	r.db.reportMutex.Lock()
	defer r.db.reportMutex.Unlock()
	
	report.ID = r.db.nextReportID
	r.db.nextReportID++
	
	r.db.reports[report.ID] = report
	return nil
}

// GetByID gets a report by ID
func (r *InMemoryReportRepository) GetByID(id uint64) (*domain.Report, error) {
	r.db.reportMutex.RLock()
	defer r.db.reportMutex.RUnlock()
	
	report, exists := r.db.reports[id]
	if !exists {
		return nil, domain.ErrReportNotFound
	}
	
	return report, nil
}

// GetAll gets all reports
func (r *InMemoryReportRepository) GetAll() []*domain.Report {
	r.db.reportMutex.RLock()
	defer r.db.reportMutex.RUnlock()
	
	reports := make([]*domain.Report, 0, len(r.db.reports))
	for _, report := range r.db.reports {
		reports = append(reports, report)
	}
	
	return reports
}

// GetByStatus gets reports by status
func (r *InMemoryReportRepository) GetByStatus(status domain.ReportStatus) []*domain.Report {
	r.db.reportMutex.RLock()
	defer r.db.reportMutex.RUnlock()
	
	var reports []*domain.Report
	for _, report := range r.db.reports {
		if report.Status == status {
			reports = append(reports, report)
		}
	}
	
	return reports
}

// Update updates a report
func (r *InMemoryReportRepository) Update(report *domain.Report) error {
	r.db.reportMutex.Lock()
	defer r.db.reportMutex.Unlock()
	
	if _, exists := r.db.reports[report.ID]; !exists {
		return domain.ErrReportNotFound
	}
	
	r.db.reports[report.ID] = report
	return nil
}
