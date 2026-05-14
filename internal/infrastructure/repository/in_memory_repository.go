package repository

import (
	"errors"
	"sync"

	"twitter-clone/internal/domain"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user domain.User) error
	GetByID(id int64) (domain.User, error)
	GetByUsername(username string) (domain.User, error)
	Update(user domain.User) error
	Delete(id int64) error
	GetAll() []domain.User
	SearchByUsername(query string) []domain.User
}

// PostRepository defines the interface for post data access
type PostRepository interface {
	Create(post *domain.Post) error
	GetByID(id int64) (*domain.Post, error)
	Update(post *domain.Post) error
	Delete(id int64) error
	GetByAuthorID(authorID int64) []*domain.Post
	GetAll() []*domain.Post
	GetByHashtag(hashtagID int64) []*domain.Post
	SearchByContent(query string) []*domain.Post
	GetPopular(limit int) []*domain.Post
	GetThread(parentID int64) []*domain.Post
}

// HashtagRepository defines the interface for hashtag data access
type HashtagRepository interface {
	Create(hashtag *domain.Hashtag) error
	GetByID(id int64) (*domain.Hashtag, error)
	GetByTitle(title string) (*domain.Hashtag, error)
	Update(hashtag *domain.Hashtag) error
	GetAll() []*domain.Hashtag
	GetPopular(limit int) []*domain.Hashtag
	Search(query string) []*domain.Hashtag
}

// ReportRepository defines the interface for report data access
type ReportRepository interface {
	Create(report *domain.Report) error
	GetByID(id int64) (*domain.Report, error)
	Update(report *domain.Report) error
	GetAll() []*domain.Report
	GetByStatus(status domain.ReportStatus) []*domain.Report
}

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	Create(notification *domain.Notification) error
	GetByUserID(userID int64) []*domain.Notification
	MarkAsRead(notificationID int64) error
	GetUnreadCount(userID int64) int
}

// BookmarkRepository defines the interface for bookmark data access
type BookmarkRepository interface {
	AddBookmark(userID, postID int64) error
	RemoveBookmark(userID, postID int64) error
	GetByUserID(userID int64) []int64
}

// InMemoryUserRepository is an in-memory implementation of UserRepository
type InMemoryUserRepository struct {
	users      map[int64]domain.User
	usernameMap map[string]int64
	mu         sync.RWMutex
	nextID     int64
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:       make(map[int64]domain.User),
		usernameMap: make(map[string]int64),
		nextID:      1,
	}
}

func (r *InMemoryUserRepository) Create(user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usernameMap[user.GetUsername()]; exists {
		return errors.New("username already exists")
	}

	id := r.nextID
	r.nextID++

	// Set ID using reflection or type assertion
	switch u := user.(type) {
	case *domain.NormalUser:
		u.ID = id
		r.users[id] = u
	case *domain.BlueUser:
		u.ID = id
		r.users[id] = u
	case *domain.GoldUser:
		u.ID = id
		r.users[id] = u
	}

	r.usernameMap[user.GetUsername()] = id
	return nil
}

func (r *InMemoryUserRepository) GetByID(id int64) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepository) GetByUsername(username string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.usernameMap[username]
	if !exists {
		return nil, errors.New("user not found")
	}

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *InMemoryUserRepository) Update(user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.GetID()]; !exists {
		return errors.New("user not found")
	}

	r.users[user.GetID()] = user
	return nil
}

func (r *InMemoryUserRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return errors.New("user not found")
	}

	delete(r.usernameMap, user.GetUsername())
	delete(r.users, id)
	return nil
}

func (r *InMemoryUserRepository) GetAll() []domain.User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users
}

func (r *InMemoryUserRepository) SearchByUsername(query string) []domain.User {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []domain.User
	for _, user := range r.users {
		if containsIgnoreCase(user.GetUsername(), query) {
			results = append(results, user)
		}
	}
	return results
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

// InMemoryPostRepository is an in-memory implementation of PostRepository
type InMemoryPostRepository struct {
	posts  map[int64]*domain.Post
	mu     sync.RWMutex
	nextID int64
}

func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		posts:  make(map[int64]*domain.Post),
		nextID: 1,
	}
}

func (r *InMemoryPostRepository) Create(post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post.ID = r.nextID
	r.nextID++
	post.ShareURL = "https://x.com/Posts/" + formatInt64(post.ID)
	r.posts[post.ID] = post
	return nil
}

func (r *InMemoryPostRepository) GetByID(id int64) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.posts[id]
	if !exists {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (r *InMemoryPostRepository) Update(post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[post.ID]; !exists {
		return errors.New("post not found")
	}

	r.posts[post.ID] = post
	return nil
}

func (r *InMemoryPostRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, exists := r.posts[id]
	if !exists {
		return errors.New("post not found")
	}

	post.Deleted = true
	r.posts[id] = post
	return nil
}

func (r *InMemoryPostRepository) GetByAuthorID(authorID int64) []*domain.Post {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var posts []*domain.Post
	for _, post := range r.posts {
		if post.AuthorID == authorID && !post.Deleted {
			posts = append(posts, post)
		}
	}
	return posts
}

func (r *InMemoryPostRepository) GetAll() []*domain.Post {
	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]*domain.Post, 0, len(r.posts))
	for _, post := range r.posts {
		if !post.Deleted && !post.Locked {
			posts = append(posts, post)
		}
	}
	return posts
}

func (r *InMemoryPostRepository) GetByHashtag(hashtagID int64) []*domain.Post {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var posts []*domain.Post
	for _, post := range r.posts {
		if !post.Deleted && !post.Locked {
			for _, tag := range post.Hashtags {
				if tag.ID == hashtagID {
					posts = append(posts, post)
					break
				}
			}
		}
	}
	return posts
}

func (r *InMemoryPostRepository) SearchByContent(query string) []*domain.Post {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var posts []*domain.Post
	for _, post := range r.posts {
		if !post.Deleted && !post.Locked && containsIgnoreCase(post.Content, query) {
			posts = append(posts, post)
		}
	}
	return posts
}

func (r *InMemoryPostRepository) GetPopular(limit int) []*domain.Post {
	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]*domain.Post, 0)
	for _, post := range r.posts {
		if !post.Deleted && !post.Locked {
			posts = append(posts, post)
		}
	}

	// Sort by like count
	for i := 0; i < len(posts); i++ {
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

func (r *InMemoryPostRepository) GetThread(parentID int64) []*domain.Post {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var thread []*domain.Post
	for _, post := range r.posts {
		if !post.Deleted && post.ParentPostID != nil && *post.ParentPostID == parentID {
			thread = append(thread, post)
		}
	}
	return thread
}

// InMemoryHashtagRepository is an in-memory implementation
type InMemoryHashtagRepository struct {
	hashtags map[int64]*domain.Hashtag
	titleMap map[string]int64
	mu       sync.RWMutex
	nextID   int64
}

func NewInMemoryHashtagRepository() *InMemoryHashtagRepository {
	return &InMemoryHashtagRepository{
		hashtags: make(map[int64]*domain.Hashtag),
		titleMap: make(map[string]int64),
		nextID:   1,
	}
}

func (r *InMemoryHashtagRepository) Create(hashtag *domain.Hashtag) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.titleMap[hashtag.Title]; exists {
		return errors.New("hashtag already exists")
	}

	hashtag.ID = r.nextID
	r.nextID++
	r.hashtags[hashtag.ID] = hashtag
	r.titleMap[hashtag.Title] = hashtag.ID
	return nil
}

func (r *InMemoryHashtagRepository) GetByID(id int64) (*domain.Hashtag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hashtag, exists := r.hashtags[id]
	if !exists {
		return nil, errors.New("hashtag not found")
	}
	return hashtag, nil
}

func (r *InMemoryHashtagRepository) GetByTitle(title string) (*domain.Hashtag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.titleMap[title]
	if !exists {
		return nil, errors.New("hashtag not found")
	}

	hashtag, exists := r.hashtags[id]
	if !exists {
		return nil, errors.New("hashtag not found")
	}
	return hashtag, nil
}

func (r *InMemoryHashtagRepository) Update(hashtag *domain.Hashtag) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.hashtags[hashtag.ID]; !exists {
		return errors.New("hashtag not found")
	}

	r.hashtags[hashtag.ID] = hashtag
	return nil
}

func (r *InMemoryHashtagRepository) GetAll() []*domain.Hashtag {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make([]*domain.Hashtag, 0, len(r.hashtags))
	for _, tag := range r.hashtags {
		tags = append(tags, tag)
	}
	return tags
}

func (r *InMemoryHashtagRepository) GetPopular(limit int) []*domain.Hashtag {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tags := make([]*domain.Hashtag, 0, len(r.hashtags))
	for _, tag := range r.hashtags {
		tags = append(tags, tag)
	}

	// Sort by post count
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			if len(tags[j].Posts) > len(tags[i].Posts) {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	if limit > len(tags) {
		limit = len(tags)
	}
	return tags[:limit]
}

func (r *InMemoryHashtagRepository) Search(query string) []*domain.Hashtag {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.Hashtag
	for _, tag := range r.hashtags {
		if containsIgnoreCase(tag.Title, query) {
			results = append(results, tag)
		}
	}
	return results
}

// InMemoryReportRepository
type InMemoryReportRepository struct {
	reports  map[int64]*domain.Report
	mu       sync.RWMutex
	nextID   int64
}

func NewInMemoryReportRepository() *InMemoryReportRepository {
	return &InMemoryReportRepository{
		reports: make(map[int64]*domain.Report),
		nextID:  1,
	}
}

func (r *InMemoryReportRepository) Create(report *domain.Report) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	report.ID = r.nextID
	r.nextID++
	r.reports[report.ID] = report
	return nil
}

func (r *InMemoryReportRepository) GetByID(id int64) (*domain.Report, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	report, exists := r.reports[id]
	if !exists {
		return nil, errors.New("report not found")
	}
	return report, nil
}

func (r *InMemoryReportRepository) Update(report *domain.Report) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.reports[report.ID]; !exists {
		return errors.New("report not found")
	}

	r.reports[report.ID] = report
	return nil
}

func (r *InMemoryReportRepository) GetAll() []*domain.Report {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reports := make([]*domain.Report, 0, len(r.reports))
	for _, report := range r.reports {
		reports = append(reports, report)
	}
	return reports
}

func (r *InMemoryReportRepository) GetByStatus(status domain.ReportStatus) []*domain.Report {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var reports []*domain.Report
	for _, report := range r.reports {
		if report.Status == status {
			reports = append(reports, report)
		}
	}
	return reports
}

// InMemoryNotificationRepository
type InMemoryNotificationRepository struct {
	notifications map[int64][]*domain.Notification
	mu            sync.RWMutex
	nextID        int64
}

func NewInMemoryNotificationRepository() *InMemoryNotificationRepository {
	return &InMemoryNotificationRepository{
		notifications: make(map[int64][]*domain.Notification),
		nextID:        1,
	}
}

func (r *InMemoryNotificationRepository) Create(notification *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	notification.ID = r.nextID
	r.nextID++
	r.notifications[notification.UserID] = append(r.notifications[notification.UserID], notification)
	return nil
}

func (r *InMemoryNotificationRepository) GetByUserID(userID int64) []*domain.Notification {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.notifications[userID]
}

func (r *InMemoryNotificationRepository) MarkAsRead(notificationID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, notifications := range r.notifications {
		for _, notif := range notifications {
			if notif.ID == notificationID {
				notif.Read = true
				return nil
			}
		}
	}
	return errors.New("notification not found")
}

func (r *InMemoryNotificationRepository) GetUnreadCount(userID int64) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, notif := range r.notifications[userID] {
		if !notif.Read {
			count++
		}
	}
	return count
}

// InMemoryBookmarkRepository
type InMemoryBookmarkRepository struct {
	bookmarks map[int64]map[int64]bool
	mu        sync.RWMutex
}

func NewInMemoryBookmarkRepository() *InMemoryBookmarkRepository {
	return &InMemoryBookmarkRepository{
		bookmarks: make(map[int64]map[int64]bool),
	}
}

func (r *InMemoryBookmarkRepository) AddBookmark(userID, postID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bookmarks[userID]; !exists {
		r.bookmarks[userID] = make(map[int64]bool)
	}
	r.bookmarks[userID][postID] = true
	return nil
}

func (r *InMemoryBookmarkRepository) RemoveBookmark(userID, postID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if userBookmarks, exists := r.bookmarks[userID]; exists {
		delete(userBookmarks, postID)
	}
	return nil
}

func (r *InMemoryBookmarkRepository) GetByUserID(userID int64) []int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var postIDs []int64
	if userBookmarks, exists := r.bookmarks[userID]; exists {
		for postID := range userBookmarks {
			postIDs = append(postIDs, postID)
		}
	}
	return postIDs
}

// Helper function
func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	
	negative := n < 0
	if negative {
		n = -n
	}
	
	digits := make([]byte, 0)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	
	return string(digits)
}

// Database Singleton
type Database struct {
	UserRepo       UserRepository
	PostRepo       PostRepository
	HashtagRepo    HashtagRepository
	ReportRepo     ReportRepository
	NotifRepo      NotificationRepository
	BookmarkRepo   BookmarkRepository
	initialized    bool
	mu             sync.Mutex
}

var dbInstance *Database
var dbOnce sync.Once

func GetDatabaseInstance() *Database {
	dbOnce.Do(func() {
		dbInstance = &Database{
			UserRepo:       NewInMemoryUserRepository(),
			PostRepo:       NewInMemoryPostRepository(),
			HashtagRepo:    NewInMemoryHashtagRepository(),
			ReportRepo:     NewInMemoryReportRepository(),
			NotifRepo:      NewInMemoryNotificationRepository(),
			BookmarkRepo:   NewInMemoryBookmarkRepository(),
			initialized:    true,
		}
	})
	return dbInstance
}

func (d *Database) Initialize() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if !d.initialized {
		d.UserRepo = NewInMemoryUserRepository()
		d.PostRepo = NewInMemoryPostRepository()
		d.HashtagRepo = NewInMemoryHashtagRepository()
		d.ReportRepo = NewInMemoryReportRepository()
		d.NotifRepo = NewInMemoryNotificationRepository()
		d.BookmarkRepo = NewInMemoryBookmarkRepository()
		d.initialized = true
	}
}
