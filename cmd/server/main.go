package main

import (
	"fmt"
	"log"
	"net/http"

	httpHandler "twitter-clone/internal/interface/http"
	"twitter-clone/internal/infrastructure/repository"
	"twitter-clone/internal/usecase"
)

func main() {
	// Initialize database (Singleton)
	db := repository.GetDatabaseInstance()

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(db.UserRepo)
	postUseCase := usecase.NewPostUseCase(db.PostRepo, db.UserRepo, db.HashtagRepo, db.NotifRepo)
	userUseCase := usecase.NewUserUseCase(db.UserRepo, db.PostRepo, db.NotifRepo)
	adminUseCase := usecase.NewAdminUseCase(db.UserRepo, db.PostRepo, db.ReportRepo)

	// Initialize handlers
	authHandler := httpHandler.NewAuthHandler(authUseCase)
	postHandler := httpHandler.NewPostHandler(postUseCase)
	userHandler := httpHandler.NewUserHandler(userUseCase)
	adminHandler := httpHandler.NewAdminHandler(adminUseCase)

	// Setup routes
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/login", authHandler.Login)

	// Post routes
	mux.HandleFunc("/api/posts", postHandler.CreatePost)
	mux.HandleFunc("/api/posts/popular", postHandler.GetPopularPosts)
	mux.HandleFunc("/api/posts/search", postHandler.SearchPosts)
	mux.HandleFunc("/api/posts/view", postHandler.ViewPost)
	mux.HandleFunc("/api/posts/like", postHandler.LikePost)
	mux.HandleFunc("/api/posts/edit", postHandler.EditPost)
	mux.HandleFunc("/api/posts/delete", postHandler.DeletePost)
	mux.HandleFunc("/api/posts/thread", postHandler.GetThread)

	// User routes
	mux.HandleFunc("/api/users/follow", userHandler.FollowUser)
	mux.HandleFunc("/api/users/unfollow", userHandler.UnfollowUser)
	mux.HandleFunc("/api/users/upgrade", userHandler.UpgradeSubscription)
	mux.HandleFunc("/api/users/tokens", userHandler.AddTokens)
	mux.HandleFunc("/api/users/profile", userHandler.GetUserProfile)
	mux.HandleFunc("/api/users/search", userHandler.SearchUsers)
	mux.HandleFunc("/api/users/notifications", userHandler.GetNotifications)

	// Admin routes
	mux.HandleFunc("/api/admin/users/block", adminHandler.BlockUser)
	mux.HandleFunc("/api/admin/users/unblock", adminHandler.UnblockUser)
	mux.HandleFunc("/api/admin/posts/block", adminHandler.BlockPost)
	mux.HandleFunc("/api/admin/reports/create", adminHandler.CreateReport)
	mux.HandleFunc("/api/admin/reports/confirm", adminHandler.ConfirmReport)
	mux.HandleFunc("/api/admin/reports/reject", adminHandler.RejectReport)
	mux.HandleFunc("/api/admin/reports/all", adminHandler.GetAllReports)
	mux.HandleFunc("/api/admin/reports/pending", adminHandler.GetPendingReports)

	// Start server
	fmt.Println("Twitter Clone Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
