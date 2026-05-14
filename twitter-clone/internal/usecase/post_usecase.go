package usecase

import (
	"regexp"
	"strings"
	"twitter-clone/internal/domain"
	"twitter-clone/internal/infrastructure/repository"
)

// PostUseCase handles post business logic
type PostUseCase struct {
	postRepo   repository.PostRepository
	userRepo   repository.UserRepository
	hashtagRepo repository.HashtagRepository
}

// NewPostUseCase creates a new PostUseCase
func NewPostUseCase(
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
	hashtagRepo repository.HashtagRepository,
) *PostUseCase {
	return &PostUseCase{
		postRepo:    postRepo,
		userRepo:    userRepo,
		hashtagRepo: hashtagRepo,
	}
}

// CreatePost creates a new post
func (uc *PostUseCase) CreatePost(userID uint64, req domain.CreatePostRequest) (*domain.Post, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user.Blocked {
		return nil, domain.ErrUserBlocked
	}

	// Calculate cost based on user type
	var cost int64
	hasMedia := req.MediaType != "" && req.MediaPath != ""
	
	// Type assertion to get the correct user type for cost calculation
	if user.Badge == "gold" {
		goldUser := &domain.GoldUser{PremiumUser: domain.PremiumUser{User: *user}}
		cost = goldUser.CalculatePostCost(len(req.Text), hasMedia)
	} else if user.Badge == "blue" {
		blueUser := &domain.BlueUser{PremiumUser: domain.PremiumUser{User: *user}}
		cost = blueUser.CalculatePostCost(len(req.Text), hasMedia)
	} else {
		normalUser := &domain.NormalUser{User: *user}
		cost = normalUser.CalculatePostCost(len(req.Text), hasMedia)
	}

	// Check if user has enough tokens
	if user.Token < cost {
		return nil, domain.ErrInsufficientTokens
	}

	// Deduct tokens
	user.Token -= cost

	// Create post
	post := domain.NewPost(0, req.Text, userID)

	// Add media if provided
	if hasMedia {
		if req.MediaType == "image" {
			image := domain.NewImage(0, req.MediaPath, "JPEG")
			post.Media = image.File
		} else if req.MediaType == "video" {
			video := domain.NewVideo(0, req.MediaPath, "MP4", "720", 30)
			post.Media = video.File
		}
	}

	// Extract and add hashtags
	hashtagStrings := domain.ExtractHashtags(req.Text)
	for _, tagStr := range hashtagStrings {
		hashtag, err := uc.hashtagRepo.GetByTitle(tagStr)
		if err != nil {
			// Create new hashtag
			hashtag = domain.NewHashtag(0, tagStr)
			if err := uc.hashtagRepo.Create(hashtag); err != nil {
				continue
			}
		}
		
		post.AddHashtag(*hashtag)
		hashtag.AddPost(post.ID)
		uc.hashtagRepo.Update(hashtag)
	}

	// Save post
	if err := uc.postRepo.Create(post); err != nil {
		return nil, err
	}

	// Add post to user's post list
	user.PostIDs = append(user.PostIDs, post.ID)
	if err := uc.userRepo.Update(user); err != nil {
		return nil, err
	}

	return post, nil
}

// GetPost gets a post by ID
func (uc *PostUseCase) GetPost(postID uint64) (*domain.Post, error) {
	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}

	if post.Deleted {
		return nil, domain.ErrPostDeleted
	}

	if post.Locked {
		return nil, domain.ErrPostLocked
	}

	// Increment view count
	post.IncrementView()
	uc.postRepo.Update(post)

	return post, nil
}

// GetPopularPosts gets popular posts
func (uc *PostUseCase) GetPopularPosts(limit int) []*domain.Post {
	return uc.postRepo.GetPopular(limit)
}

// GetUserPosts gets posts by a specific user
func (uc *PostUseCase) GetUserPosts(userID uint64) []*domain.Post {
	return uc.postRepo.GetByAuthorID(userID)
}

// GetAllPosts gets all posts
func (uc *PostUseCase) GetAllPosts() []*domain.Post {
	return uc.postRepo.GetAll()
}

// LikePost likes a post
func (uc *PostUseCase) LikePost(userID, postID uint64) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return err
	}

	if post.AuthorID == userID {
		return domain.ErrCannotLikeOwnPost
	}

	if post.Deleted || post.Locked {
		return domain.ErrPostDeleted
	}

	// Check if already liked
	if post.IsLikedBy(userID) {
		// Unlike
		post.Unlike(userID)
		
		// Remove from user's liked posts
		for i, id := range user.LikedPostIDs {
			if id == postID {
				user.LikedPostIDs = append(user.LikedPostIDs[:i], user.LikedPostIDs[i+1:]...)
				break
			}
		}
	} else {
		// Like
		post.Like(userID)
		user.LikedPostIDs = append(user.LikedPostIDs, postID)
	}

	// Update post and user
	if err := uc.postRepo.Update(post); err != nil {
		return err
	}

	return uc.userRepo.Update(user)
}

// DeletePost deletes a post
func (uc *PostUseCase) DeletePost(userID, postID uint64) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return err
	}

	// Only author can delete
	if post.AuthorID != userID {
		return domain.ErrInvalidCredentials
	}

	// Mark as deleted
	post.MarkAsDeleted()

	// Remove from user's post list
	for i, id := range user.PostIDs {
		if id == postID {
			user.PostIDs = append(user.PostIDs[:i], user.PostIDs[i+1:]...)
			break
		}
	}

	// Update post and user
	if err := uc.postRepo.Update(post); err != nil {
		return err
	}

	return uc.userRepo.Update(user)
}

// EditPost edits a post (premium users only)
func (uc *PostUseCase) EditPost(userID, postID uint64, newText string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Only premium users can edit
	if !user.IsPremium() {
		return domain.ErrInvalidCredentials
	}

	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return err
	}

	// Only author can edit
	if post.AuthorID != userID {
		return domain.ErrInvalidCredentials
	}

	if post.Deleted || post.Locked {
		return domain.ErrPostDeleted
	}

	post.Text = newText
	post.MarkAsEdited()

	return uc.postRepo.Update(post)
}

// ReplyToPost creates a reply to a post
func (uc *PostUseCase) ReplyToPost(userID, parentPostID uint64, text string) (*domain.Post, error) {
	parentPost, err := uc.postRepo.GetByID(parentPostID)
	if err != nil {
		return nil, err
	}

	if parentPost.Deleted || parentPost.Locked {
		return nil, domain.ErrPostDeleted
	}

	// Check max replies
	if len(parentPost.ReplyPostIDs) >= 10000 {
		return nil, domain.ErrMaxRepliesReached
	}

	// Create reply post
	replyReq := domain.CreatePostRequest{
		Text: text,
	}
	
	replyPost, err := uc.CreatePost(userID, replyReq)
	if err != nil {
		return nil, err
	}

	// Set parent post ID
	replyPost.ParentPostID = &parentPostID

	// Add reply to parent post
	parentPost.AddReply(replyPost.ID)

	// Get author username for the reply prefix
	author, err := uc.userRepo.GetByID(parentPost.AuthorID)
	if err == nil {
		replyPost.Text = "در پاسخ به @" + author.Username + ": " + text
	}

	// Update parent post
	if err := uc.postRepo.Update(parentPost); err != nil {
		return nil, err
	}

	// Update reply post with parent info
	if err := uc.postRepo.Update(replyPost); err != nil {
		return nil, err
	}

	return replyPost, nil
}

// SearchPosts searches posts by text
func (uc *PostUseCase) SearchPosts(query string) []*domain.Post {
	return uc.postRepo.SearchByText(query)
}

// GenerateShareLink generates a share link for a post
func (uc *PostUseCase) GenerateShareLink(postID uint64) string {
	return "https://x.com/Posts/" + string(rune(postID))
}

// GetReplies gets all replies to a post
func (uc *PostUseCase) GetReplies(postID uint64) []*domain.Post {
	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return nil
	}

	var replies []*domain.Post
	for _, replyID := range post.ReplyPostIDs {
		reply, err := uc.postRepo.GetByID(replyID)
		if err == nil && !reply.Deleted {
			replies = append(replies, reply)
		}
	}

	return replies
}

// BlockPost blocks a post (admin function)
func (uc *PostUseCase) BlockPost(adminID, postID uint64) error {
	// Verify admin
	admin := domain.GetAdminInstance()
	if admin.ID != adminID {
		return domain.ErrInvalidCredentials
	}

	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return err
	}

	post.Locked = true
	return uc.postRepo.Update(post)
}

// UnblockPost unblocks a post (admin function)
func (uc *PostUseCase) UnblockPost(adminID, postID uint64) error {
	// Verify admin
	admin := domain.GetAdminInstance()
	if admin.ID != adminID {
		return domain.ErrInvalidCredentials
	}

	post, err := uc.postRepo.GetByID(postID)
	if err != nil {
		return err
	}

	post.Locked = false
	return uc.postRepo.Update(post)
}

// RecommendPosts recommends posts to a user based on their interests
func (uc *PostUseCase) RecommendPosts(userID uint64, limit int) []*domain.Post {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		return []*domain.Post{}
	}

	allPosts := uc.postRepo.GetAll()
	
	// Simple recommendation based on liked posts and followed users
	// In production, you'd use more sophisticated algorithms
	
	var recommended []*domain.Post
	scoreMap := make(map[uint64]int)

	// Score posts from followed users
	for _, followingID := range user.FollowingIDs {
		for _, post := range allPosts {
			if post.AuthorID == followingID && !post.Deleted && !post.Locked {
				scoreMap[post.ID] += 2
			}
		}
	}

	// Score posts with similar hashtags to liked posts
	for _, likedPostID := range user.LikedPostIDs {
		likedPost, err := uc.postRepo.GetByID(likedPostID)
		if err != nil {
			continue
		}

		for _, hashtag := range likedPost.Hashtags {
			for _, post := range allPosts {
				if post.ID != likedPostID && !post.Deleted && !post.Locked {
					for _, postHashtag := range post.Hashtags {
						if postHashtag.Title == hashtag.Title {
							scoreMap[post.ID]++
						}
					}
				}
			}
		}
	}

	// Sort by score and return top recommendations
	type scoredPost struct {
		post  *domain.Post
		score int
	}

	var scoredPosts []scoredPost
	for id, score := range scoreMap {
		post, err := uc.postRepo.GetByID(id)
		if err == nil && !post.Deleted && !post.Locked {
			scoredPosts = append(scoredPosts, scoredPost{post: post, score: score})
		}
	}

	// Sort by score
	for i := 0; i < len(scoredPosts)-1; i++ {
		for j := i + 1; j < len(scoredPosts); j++ {
			if scoredPosts[j].score > scoredPosts[i].score {
				scoredPosts[i], scoredPosts[j] = scoredPosts[j], scoredPosts[i]
			}
		}
	}

	// Return top recommendations
	for i, sp := range scoredPosts {
		if i >= limit {
			break
		}
		recommended = append(recommended, sp.post)
	}

	return recommended
}

// ReportPost creates a report for a post
func (uc *PostUseCase) ReportPost(reporterID, postID, reportedUserID uint64, description string) (*domain.Report, error) {
	_, err := uc.userRepo.GetByID(reporterID)
	if err != nil {
		return nil, err
	}

	_, err = uc.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}

	report := domain.NewReport(0, reporterID, postID, reportedUserID, description)
	
	// Save report (would need report repo injected)
	// For now, just return the report
	return report, nil
}

// SortPosts sorts posts by criteria
func (uc *PostUseCase) SortPosts(posts []*domain.Post, sortBy string) []*domain.Post {
	sorted := make([]*domain.Post, len(posts))
	copy(sorted, posts)

	switch sortBy {
	case "likes":
		// Sort by likes (descending)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].LikeCount > sorted[i].LikeCount {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	case "views":
		// Sort by views (descending)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].ViewCount > sorted[i].ViewCount {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	case "date":
		// Sort by date (newest first)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].CreatedAt.After(sorted[i].CreatedAt) {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}

	return sorted
}

// FilterPostsByDate filters posts by date range
func (uc *PostUseCase) FilterPostsByDate(posts []*domain.Post, startDate, endDate string) []*domain.Post {
	// Parse dates (simplified - in production use proper date parsing)
	var filtered []*domain.Post
	
	for _, post := range posts {
		postDate := post.CreatedAt.Format("2006-01-02")
		if postDate >= startDate && postDate <= endDate {
			filtered = append(filtered, post)
		}
	}
	
	return filtered
}

// ExtractHashtagsFromText extracts hashtags from text
func (uc *PostUseCase) ExtractHashtagsFromText(text string) []string {
	hashtagRegex := regexp.MustCompile(`#\w+`)
	matches := hashtagRegex.FindAllString(text, -1)
	
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, strings.ToLower(match))
	}
	
	return result
}
