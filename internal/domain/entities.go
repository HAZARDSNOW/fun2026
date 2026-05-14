package domain

import "time"

// File represents media files (Abstract)
type File struct {
	Path string `json:"path"`
}

// Image represents an image file
type Image struct {
	File
	Format string `json:"format"` // JPEG, PNG, SVG, WEBP
}

// Video represents a video file
type Video struct {
	File
	Quality   string `json:"quality"` // 360, 720, 1080
	Format    string `json:"format"`  // MP4, MKV, MOV, WMV
	Duration  int    `json:"duration"` // in seconds
}

// Post represents a tweet/post
type Post struct {
	ID            int64     `json:"id"`
	Content       string    `json:"content"`
	Media         File      `json:"media,omitempty"`
	AuthorID      int64     `json:"author_id"`
	CreatedAt     time.Time `json:"created_at"`
	Hashtags      []Hashtag `json:"hashtags"`
	LikedBy       []int64   `json:"liked_by"`
	ParentPostID  *int64    `json:"parent_post_id,omitempty"`
	Replies       []int64   `json:"replies"`
	ViewCount     int       `json:"view_count"`
	LikeCount     int       `json:"like_count"`
	Locked        bool      `json:"locked"`
	Deleted       bool      `json:"deleted"`
	Edited        bool      `json:"edited"`
	ShareURL      string    `json:"share_url"`
}

// Hashtag represents a hashtag
type Hashtag struct {
	ID    int64   `json:"id"`
	Title string  `json:"title"`
	Posts []int64 `json:"posts"`
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
	ID              int64        `json:"id"`
	ReporterID      int64        `json:"reporter_id"`
	ReportedContent int64        `json:"reported_content"` // Post ID or User ID
	ReportedUser    int64        `json:"reported_user"`
	Status          ReportStatus `json:"status"`
	Description     string       `json:"description"`
}

// Notification represents a user notification
type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"` // like, reply, follow, mention
	Content   string    `json:"content"`
	FromUser  int64     `json:"from_user"`
	PostID    *int64    `json:"post_id,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// Bookmark represents a saved post
type Bookmark struct {
	UserID int64 `json:"user_id"`
	PostID int64 `json:"post_id"`
}

// TrendingTopic represents trending topics
type TrendingTopic struct {
	Hashtag     string `json:"hashtag"`
	TweetCount  int    `json:"tweet_count"`
	Category    string `json:"category"`
}

// Analytics represents post/user analytics
type Analytics struct {
	PostID      int64 `json:"post_id,omitempty"`
	UserID      int64 `json:"user_id,omitempty"`
	Impressions int   `json:"impressions"`
	Engagements int   `json:"engagements"`
	Likes       int   `json:"likes"`
	Replies     int   `json:"replies"`
	Shares      int   `json:"shares"`
}
