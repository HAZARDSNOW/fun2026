package domain

import (
	"time"
)

// File represents the base file entity (cannot be instantiated directly)
type File struct {
	ID       uint64 `json:"id"`
	FilePath string `json:"file_path"`
}

// Image represents an image file
type Image struct {
	File
	Format string `json:"format"` // JPEG, PNG, SVG, WEBP
}

// Video represents a video file
type Video struct {
	File
	Quality   string `json:"quality"`   // 360, 720, 1080
	Format    string `json:"format"`    // MP4, MKV, MOV, WMV
	Duration  int    `json:"duration"`  // in seconds
}

// Post represents a post/tweet
type Post struct {
	ID            uint64    `json:"id"`
	Text          string    `json:"text"`
	Media         File      `json:"media,omitempty"`
	AuthorID      uint64    `json:"author_id"`
	CreatedAt     time.Time `json:"created_at"`
	Hashtags      []Hashtag `json:"hashtags"`
	LikedByUserIDs []uint64  `json:"liked_by_user_ids"`
	ParentPostID  *uint64   `json:"parent_post_id,omitempty"`
	ReplyPostIDs  []uint64  `json:"reply_post_ids"`
	ViewCount     int64     `json:"view_count"`
	LikeCount     int64     `json:"like_count"`
	Locked        bool      `json:"locked"`
	Deleted       bool      `json:"deleted"`
	Edited        bool      `json:"edited"`
}

// Hashtag represents a hashtag
type Hashtag struct {
	ID     uint64   `json:"id"`
	Title  string   `json:"title"`
	PostIDs []uint64 `json:"post_ids"`
}

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportWaiting   ReportStatus = "WAITING"
	ReportConfirmed ReportStatus = "CONFIRMED"
	ReportRejected  ReportStatus = "REJECTED"
)

// Report represents a user report
type Report struct {
	ID              uint64       `json:"id"`
	ReporterUserID  uint64       `json:"reporter_user_id"`
	ContentID       uint64       `json:"content_id"`
	ReportedUserID  uint64       `json:"reported_user_id"`
	Status          ReportStatus `json:"status"`
	Description     string       `json:"description"`
	CreatedAt       time.Time    `json:"created_at"`
}

// NewImage creates a new image
func NewImage(id uint64, filePath, format string) *Image {
	return &Image{
		File: File{
			ID:       id,
			FilePath: filePath,
		},
		Format: format,
	}
}

// NewVideo creates a new video
func NewVideo(id uint64, filePath, format, quality string, duration int) *Video {
	return &Video{
		File: File{
			ID:       id,
			FilePath: filePath,
		},
		Quality:  quality,
		Format:   format,
		Duration: duration,
	}
}

// NewPost creates a new post
func NewPost(id uint64, text string, authorID uint64) *Post {
	return &Post{
		ID:            id,
		Text:          text,
		AuthorID:      authorID,
		CreatedAt:     time.Now(),
		Hashtags:      []Hashtag{},
		LikedByUserIDs: []uint64{},
		ReplyPostIDs:  []uint64{},
		ViewCount:     0,
		LikeCount:     0,
		Locked:        false,
		Deleted:       false,
		Edited:        false,
	}
}

// NewHashtag creates a new hashtag
func NewHashtag(id uint64, title string) *Hashtag {
	return &Hashtag{
		ID:      id,
		Title:   title,
		PostIDs: []uint64{},
	}
}

// NewReport creates a new report
func NewReport(id uint64, reporterUserID, contentID, reportedUserID uint64, description string) *Report {
	return &Report{
		ID:             id,
		ReporterUserID: reporterUserID,
		ContentID:      contentID,
		ReportedUserID: reportedUserID,
		Status:         ReportWaiting,
		Description:    description,
		CreatedAt:      time.Now(),
	}
}

// Like adds a like to the post
func (p *Post) Like(userID uint64) {
	for _, id := range p.LikedByUserIDs {
		if id == userID {
			return // Already liked
		}
	}
	p.LikedByUserIDs = append(p.LikedByUserIDs, userID)
	p.LikeCount++
}

// Unlike removes a like from the post
func (p *Post) Unlike(userID uint64) {
	for i, id := range p.LikedByUserIDs {
		if id == userID {
			p.LikedByUserIDs = append(p.LikedByUserIDs[:i], p.LikedByUserIDs[i+1:]...)
			p.LikeCount--
			break
		}
	}
}

// IsLikedBy checks if the post is liked by a user
func (p *Post) IsLikedBy(userID uint64) bool {
	for _, id := range p.LikedByUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// IncrementView increments the view count
func (p *Post) IncrementView() {
	p.ViewCount++
}

// AddReply adds a reply post ID
func (p *Post) AddReply(replyPostID uint64) {
	if len(p.ReplyPostIDs) < 10000 {
		p.ReplyPostIDs = append(p.ReplyPostIDs, replyPostID)
	}
}

// MarkAsDeleted marks the post as deleted
func (p *Post) MarkAsDeleted() {
	p.Deleted = true
}

// MarkAsEdited marks the post as edited
func (p *Post) MarkAsEdited() {
	p.Edited = true
}

// AddHashtag adds a hashtag to the post
func (p *Post) AddHashtag(hashtag Hashtag) {
	p.Hashtags = append(p.Hashtags, hashtag)
}

// AddPostToHashtag adds a post ID to the hashtag
func (h *Hashtag) AddPost(postID uint64) {
	for _, id := range h.PostIDs {
		if id == postID {
			return // Already added
		}
	}
	h.PostIDs = append(h.PostIDs, postID)
}

// ConfirmReport confirms a report
func (r *Report) ConfirmReport() {
	r.Status = ReportConfirmed
}

// RejectReport rejects a report
func (r *Report) RejectReport() {
	r.Status = ReportRejected
}
