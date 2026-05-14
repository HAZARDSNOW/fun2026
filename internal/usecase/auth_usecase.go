package usecase

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"twitter-clone/internal/domain"
	"twitter-clone/internal/infrastructure/repository"
	"twitter-clone/pkg/hash"
	"twitter-clone/pkg/token"
)

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	userRepo repository.UserRepository
}

func NewAuthUseCase(userRepo repository.UserRepository) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo}
}

type RegisterInput struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	BirthDate string `json:"birth_date"`
	Phone     string `json:"phone"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User   domain.User `json:"user"`
	Token  string      `json:"token"`
}

func (uc *AuthUseCase) Register(input RegisterInput, interestTags []int64) (*AuthResponse, error) {
	// Validate username uniqueness
	_, err := uc.userRepo.GetByUsername(input.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(input.Email) {
		return nil, errors.New("invalid email format")
	}

	// Validate phone format (simple validation)
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	if input.Phone != "" && !phoneRegex.MatchString(input.Phone) {
		return nil, errors.New("invalid phone format")
	}

	// Validate password strength
	if len(input.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Hash password
	hashedPassword, err := hash.HashPassword(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create new user
	birthDate, _ := time.Parse("2006-01-02", input.BirthDate)
	user := &domain.NormalUser{
		Account: domain.Account{
			Username:  input.Username,
			Password:  hashedPassword,
			Email:     input.Email,
			FullName:  input.FullName,
			BirthDate: birthDate,
			Phone:     input.Phone,
			JoinDate:  time.Now(),
		},
		Balance:      0,
		Token:        1000, // Initial token bonus
		Bio:          "",
		Posts:        []int64{},
		Followers:    []int64{},
		Following:    []int64{},
		LikedPosts:   []int64{},
		Badge:        "",
		Blocked:      false,
		InterestTags: interestTags,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate JWT token
	jwtToken, err := token.GenerateToken(user.GetID(), user.GetUsername())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		User:  user,
		Token: jwtToken,
	}, nil
}

func (uc *AuthUseCase) Login(input LoginInput) (*AuthResponse, error) {
	user, err := uc.userRepo.GetByUsername(input.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is blocked
	if user.(interface{ GetBlocked() bool }).GetBlocked() {
		return nil, errors.New("account is blocked")
	}

	// Verify password
	if err := hash.VerifyPassword(user.(*domain.NormalUser).Password, input.Password); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate JWT token
	jwtToken, err := token.GenerateToken(user.GetID(), user.GetUsername())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		User:  user,
		Token: jwtToken,
	}, nil
}

// PostUseCase handles post-related business logic
type PostUseCase struct {
	postRepo       repository.PostRepository
	userRepo       repository.UserRepository
	hashtagRepo    repository.HashtagRepository
	notificationRepo repository.NotificationRepository
}

func NewPostUseCase(
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
	hashtagRepo repository.HashtagRepository,
	notificationRepo repository.NotificationRepository,
) *PostUseCase {
	return &PostUseCase{
		postRepo:       postRepo,
		userRepo:       userRepo,
		hashtagRepo:    hashtagRepo,
		notificationRepo: notificationRepo,
	}
}

type CreatePostInput struct {
	AuthorID   int64  `json:"author_id"`
	Content    string `json:"content"`
	MediaPath  string `json:"media_path,omitempty"`
	MediaType  string `json:"media_type,omitempty"` // image, video
	ParentPostID *int64 `json:"parent_post_id,omitempty"`
}

func (uc *PostUseCase) CreatePost(input CreatePostInput) (*domain.Post, error) {
	// Get user
	user, err := uc.userRepo.GetByID(input.AuthorID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Calculate cost
	hasMedia := input.MediaPath != ""
	cost := user.CalculatePostCost(len(input.Content), hasMedia)

	// Check token balance
	var tokenCount int
	var normalUser *domain.NormalUser
	
	switch u := user.(type) {
	case *domain.NormalUser:
		tokenCount = u.Token
		normalUser = u
	case *domain.BlueUser:
		tokenCount = u.Token
		normalUser = &u.NormalUser
	case *domain.GoldUser:
		tokenCount = u.Token
		normalUser = &u.NormalUser
	}

	if tokenCount < cost {
		return nil, errors.New("insufficient tokens")
	}

	// Deduct tokens
	normalUser.Token -= cost

	// Extract hashtags from content
	hashtags := uc.extractHashtags(input.Content)

	// Create media file
	var media domain.File
	if input.MediaPath != "" {
		media = domain.File{Path: input.MediaPath}
	}

	// Create post
	post := &domain.Post{
		Content:      input.Content,
		Media:        media,
		AuthorID:     input.AuthorID,
		CreatedAt:    time.Now(),
		Hashtags:     hashtags,
		LikedBy:      []int64{},
		ParentPostID: input.ParentPostID,
		Replies:      []int64{},
		ViewCount:    0,
		LikeCount:    0,
		Locked:       false,
		Deleted:      false,
		Edited:       false,
	}

	// If it's a reply, add prefix and update parent
	if input.ParentPostID != nil {
		parentPost, err := uc.postRepo.GetByID(*input.ParentPostID)
		if err != nil {
			return nil, errors.New("parent post not found")
		}

		parentUser, _ := uc.userRepo.GetByID(parentPost.AuthorID)
		post.Content = "در پاسخ به @" + parentUser.GetUsername() + ": " + post.Content
		
		// Limit replies to 10000
		if len(parentPost.Replies) >= 10000 {
			return nil, errors.New("maximum reply limit reached")
		}
		
		parentPost.Replies = append(parentPost.Replies, post.ID)
		uc.postRepo.Update(parentPost)

		// Create notification for parent post author
		notif := &domain.Notification{
			UserID:   parentPost.AuthorID,
			Type:     "reply",
			Content:  "به پست شما پاسخ داده شد",
			FromUser: input.AuthorID,
			PostID:   &post.ID,
			Read:     false,
			CreatedAt: time.Now(),
		}
		uc.notificationRepo.Create(notif)
	}

	// Save post
	if err := uc.postRepo.Create(post); err != nil {
		return nil, err
	}

	// Update user's posts
	normalUser.Posts = append(normalUser.Posts, post.ID)
	uc.userRepo.Update(user)

	// Update hashtags
	for i := range hashtags {
		tag, err := uc.hashtagRepo.GetByTitle(hashtags[i].Title)
		if err != nil {
			// Create new hashtag
			uc.hashtagRepo.Create(&hashtags[i])
		} else {
			tag.Posts = append(tag.Posts, post.ID)
			uc.hashtagRepo.Update(tag)
		}
	}

	return post, nil
}

func (uc *PostUseCase) extractHashtags(content string) []domain.Hashtag {
	var hashtags []domain.Hashtag
	words := strings.Fields(content)
	
	for _, word := range words {
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			tagTitle := word[1:]
			hashtags = append(hashtags, domain.Hashtag{
				Title: tagTitle,
				Posts: []int64{},
			})
		}
	}
	
	return hashtags
}

func (uc *PostUseCase) LikePost(postID, userID int64) error {
	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("post not found")
	}

	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if already liked
	for _, likedID := range post.LikedBy {
		if likedID == userID {
			// Unlike
			post.LikeCount--
			post.LikedBy = removeInt64(post.LikedBy, userID)
			
			// Remove from user's liked posts
			var normalUser *domain.NormalUser
			switch u := user.(type) {
			case *domain.NormalUser:
				normalUser = u
			case *domain.BlueUser:
				normalUser = &u.NormalUser
			case *domain.GoldUser:
				normalUser = &u.NormalUser
			}
			normalUser.LikedPosts = removeInt64(normalUser.LikedPosts, postID)
			uc.userRepo.Update(user)
			
			uc.postRepo.Update(post)
			return nil
		}
	}

	// Like the post
	post.LikeCount++
	post.LikedBy = append(post.LikedBy, userID)

	// Add to user's liked posts
	var normalUser *domain.NormalUser
	switch u := user.(type) {
	case *domain.NormalUser:
		normalUser = u
	case *domain.BlueUser:
		normalUser = &u.NormalUser
	case *domain.GoldUser:
		normalUser = &u.NormalUser
	}
	normalUser.LikedPosts = append(normalUser.LikedPosts, postID)
	uc.userRepo.Update(user)

	uc.postRepo.Update(post)

	// Create notification for post author
	if post.AuthorID != userID {
		notif := &domain.Notification{
			UserID:   post.AuthorID,
			Type:     "like",
			Content:  "پست شما لایک شد",
			FromUser: userID,
			PostID:   &postID,
			Read:     false,
			CreatedAt: time.Now(),
		}
		uc.notificationRepo.Create(notif)
	}

	return nil
}

func (uc *PostUseCase) ViewPost(postID int64) (*domain.Post, error) {
	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}

	// Increment view count
	post.ViewCount++
	uc.postRepo.Update(post)

	return post, nil
}

func (uc *PostUseCase) EditPost(postID, userID int64, newContent string) error {
	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("post not found")
	}

	if post.AuthorID != userID {
		return errors.New("unauthorized")
	}

	// Get user to check if premium
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !user.CanEditPost() {
		return errors.New("only premium users can edit posts")
	}

	post.Content = newContent
	post.Edited = true
	return uc.postRepo.Update(post)
}

func (uc *PostUseCase) DeletePost(postID, userID int64) error {
	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("post not found")
	}

	if post.AuthorID != userID {
		return errors.New("unauthorized")
	}

	return uc.postRepo.Delete(postID)
}

func (uc *PostUseCase) GetPopularPosts(limit int) []*domain.Post {
	return uc.postRepo.GetPopular(limit)
}

func (uc *PostUseCase) SearchPosts(query string) []*domain.Post {
	return uc.postRepo.SearchByContent(query)
}

func (uc *PostUseCase) GetThread(postID int64) []*domain.Post {
	return uc.postRepo.GetThread(postID)
}

func removeInt64(slice []int64, val int64) []int64 {
	result := make([]int64, 0)
	for _, v := range slice {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}

// UserUseCase handles user-related business logic
type UserUseCase struct {
	userRepo       repository.UserRepository
	postRepo       repository.PostRepository
	notificationRepo repository.NotificationRepository
}

func NewUserUseCase(
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	notificationRepo repository.NotificationRepository,
) *UserUseCase {
	return &UserUseCase{
		userRepo:       userRepo,
		postRepo:       postRepo,
		notificationRepo: notificationRepo,
	}
}

func (uc *UserUseCase) FollowUser(followerID, followingID int64) error {
	if followerID == followingID {
		return errors.New("cannot follow yourself")
	}

	follower, err := uc.userRepo.GetByID(followerID)
	if err != nil {
		return errors.New("follower not found")
	}

	following, err := uc.userRepo.GetByID(followingID)
	if err != nil {
		return errors.New("following not found")
	}

	// Get normal user structs
	var followerNormal *domain.NormalUser
	var followingNormal *domain.NormalUser
	
	switch u := follower.(type) {
	case *domain.NormalUser:
		followerNormal = u
	case *domain.BlueUser:
		followerNormal = &u.NormalUser
	case *domain.GoldUser:
		followerNormal = &u.NormalUser
	}
	
	switch u := following.(type) {
	case *domain.NormalUser:
		followingNormal = u
	case *domain.BlueUser:
		followingNormal = &u.NormalUser
	case *domain.GoldUser:
		followingNormal = &u.NormalUser
	}

	// Check if already following
	for _, id := range followerNormal.Following {
		if id == followingID {
			return errors.New("already following")
		}
	}

	// Add to following list
	followerNormal.Following = append(followerNormal.Following, followingID)
	
	// Add to followers list
	followingNormal.Followers = append(followingNormal.Followers, followerID)

	uc.userRepo.Update(follower)
	uc.userRepo.Update(following)

	// Create notification
	notif := &domain.Notification{
		UserID:   followingID,
		Type:     "follow",
		Content:  "شما را دنبال کرد",
		FromUser: followerID,
		Read:     false,
		CreatedAt: time.Now(),
	}
	uc.notificationRepo.Create(notif)

	return nil
}

func (uc *UserUseCase) UnfollowUser(followerID, followingID int64) error {
	follower, err := uc.userRepo.GetByID(followerID)
	if err != nil {
		return errors.New("follower not found")
	}

	following, err := uc.userRepo.GetByID(followingID)
	if err != nil {
		return errors.New("following not found")
	}

	var followerNormal *domain.NormalUser
	var followingNormal *domain.NormalUser
	
	switch u := follower.(type) {
	case *domain.NormalUser:
		followerNormal = u
	case *domain.BlueUser:
		followerNormal = &u.NormalUser
	case *domain.GoldUser:
		followerNormal = &u.NormalUser
	}
	
	switch u := following.(type) {
	case *domain.NormalUser:
		followingNormal = u
	case *domain.BlueUser:
		followingNormal = &u.NormalUser
	case *domain.GoldUser:
		followingNormal = &u.NormalUser
	}

	// Remove from following list
	followerNormal.Following = removeInt64(followerNormal.Following, followingID)
	
	// Remove from followers list
	followingNormal.Followers = removeInt64(followingNormal.Followers, followerID)

	uc.userRepo.Update(follower)
	uc.userRepo.Update(following)

	return nil
}

func (uc *UserUseCase) UpgradeSubscription(userID int64, plan string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	var normalUser *domain.NormalUser
	switch u := user.(type) {
	case *domain.NormalUser:
		normalUser = u
	case *domain.BlueUser:
		normalUser = &u.NormalUser
	case *domain.GoldUser:
		normalUser = &u.NormalUser
	}

	// Check balance (9$ for Blue, 19$ for Gold, 1000 token = 1$)
	var cost int
	var newUser domain.User

	switch plan {
	case "blue":
		cost = 9000 // 9 dollars
		if normalUser.Balance < cost {
			return errors.New("insufficient balance")
		}
		normalUser.Balance -= cost
		normalUser.Token += 3000 // Bonus tokens
		
		newUser = &domain.BlueUser{
			PremiumUser: domain.PremiumUser{
				NormalUser:      *normalUser,
				SubscriptionEnd: time.Now().AddDate(0, 1, 0), // 1 month
			},
		}
	case "gold":
		cost = 19000 // 19 dollars
		if normalUser.Balance < cost {
			return errors.New("insufficient balance")
		}
		normalUser.Balance -= cost
		normalUser.Token += 3000 // Bonus tokens
		
		newUser = &domain.GoldUser{
			PremiumUser: domain.PremiumUser{
				NormalUser:      *normalUser,
				SubscriptionEnd: time.Now().AddDate(0, 1, 0), // 1 month
			},
		}
	default:
		return errors.New("invalid plan")
	}

	return uc.userRepo.Update(newUser)
}

func (uc *UserUseCase) AddTokens(userID int64, amount int) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	var normalUser *domain.NormalUser
	switch u := user.(type) {
	case *domain.NormalUser:
		normalUser = u
	case *domain.BlueUser:
		normalUser = &u.NormalUser
	case *domain.GoldUser:
		normalUser = &u.NormalUser
	}

	// 1000 tokens = 1 dollar
	cost := amount / 1000
	if normalUser.Balance < cost*1000 {
		return errors.New("insufficient balance")
	}

	normalUser.Balance -= cost * 1000
	normalUser.Token += amount

	return uc.userRepo.Update(user)
}

func (uc *UserUseCase) GetUserProfile(userID int64, viewerID int64) (map[string]interface{}, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	var normalUser *domain.NormalUser
	switch u := user.(type) {
	case *domain.NormalUser:
		normalUser = u
	case *domain.BlueUser:
		normalUser = &u.NormalUser
	case *domain.GoldUser:
		normalUser = &u.NormalUser
	}

	// Get user's posts
	posts := uc.postRepo.GetByAuthorID(userID)

	profile := map[string]interface{}{
		"id":              user.GetID(),
		"username":        user.GetUsername(),
		"full_name":       user.(*domain.NormalUser).FullName,
		"bio":             normalUser.Bio,
		"badge":           user.GetBadge(),
		"followers_count": len(normalUser.Followers),
		"following_count": len(normalUser.Following),
		"posts_count":     len(posts),
		"posts":           posts,
	}

	return profile, nil
}

func (uc *UserUseCase) SearchUsers(query string) []domain.User {
	return uc.userRepo.SearchByUsername(query)
}

func (uc *UserUseCase) GetNotifications(userID int64) []*domain.Notification {
	return uc.notificationRepo.GetByUserID(userID)
}

// AdminUseCase handles admin operations
type AdminUseCase struct {
	userRepo   repository.UserRepository
	postRepo   repository.PostRepository
	reportRepo repository.ReportRepository
}

func NewAdminUseCase(
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	reportRepo repository.ReportRepository,
) *AdminUseCase {
	return &AdminUseCase{
		userRepo:   userRepo,
		postRepo:   postRepo,
		reportRepo: reportRepo,
	}
}

func (uc *AdminUseCase) BlockUser(userID int64) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	var normalUser *domain.NormalUser
	switch u := user.(type) {
	case *domain.NormalUser:
		normalUser = u
	case *domain.BlueUser:
		normalUser = &u.NormalUser
	case *domain.GoldUser:
		normalUser = &u.NormalUser
	}

	normalUser.Blocked = true
	return uc.userRepo.Update(user)
}

func (uc *AdminUseCase) UnblockUser(userID int64) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	var normalUser *domain.NormalUser
	switch u := user.(type) {
	case *domain.NormalUser:
		normalUser = u
	case *domain.BlueUser:
		normalUser = &u.NormalUser
	case *domain.GoldUser:
		normalUser = &u.NormalUser
	}

	normalUser.Blocked = false
	return uc.userRepo.Update(user)
}

func (uc *AdminUseCase) BlockPost(postID int64) error {
	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return errors.New("post not found")
	}

	post.Locked = true
	return uc.postRepo.Update(post)
}

func (uc *AdminUseCase) CreateReport(reporterID, reportedContent, reportedUser int64, description string) (*domain.Report, error) {
	report := &domain.Report{
		ReporterID:      reporterID,
		ReportedContent: reportedContent,
		ReportedUser:    reportedUser,
		Status:          domain.ReportWaiting,
		Description:     description,
	}

	if err := uc.reportRepo.Create(report); err != nil {
		return nil, err
	}

	return report, nil
}

func (uc *AdminUseCase) ConfirmReport(reportID int64) error {
	report, err := uc.reportRepo.GetByID(reportID)
	if err != nil {
		return errors.New("report not found")
	}

	report.Status = domain.ReportConfirmed

	// Block the reported content (assuming it's a post)
	post, err := uc.postRepo.GetByID(report.ReportedContent)
	if err == nil {
		post.Locked = true
		uc.postRepo.Update(post)
	}

	// Or block the user
	user, err := uc.userRepo.GetByID(report.ReportedUser)
	if err == nil {
		var normalUser *domain.NormalUser
		switch u := user.(type) {
		case *domain.NormalUser:
			normalUser = u
		case *domain.BlueUser:
			normalUser = &u.NormalUser
		case *domain.GoldUser:
			normalUser = &u.NormalUser
		}
		normalUser.Blocked = true
		uc.userRepo.Update(user)
	}

	return uc.reportRepo.Update(report)
}

func (uc *AdminUseCase) RejectReport(reportID int64) error {
	report, err := uc.reportRepo.GetByID(reportID)
	if err != nil {
		return errors.New("report not found")
	}

	report.Status = domain.ReportRejected
	return uc.reportRepo.Update(report)
}

func (uc *AdminUseCase) GetAllReports() []*domain.Report {
	return uc.reportRepo.GetAll()
}

func (uc *AdminUseCase) GetPendingReports() []*domain.Report {
	return uc.reportRepo.GetByStatus(domain.ReportWaiting)
}
