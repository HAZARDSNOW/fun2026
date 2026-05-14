package main

import (
	"fmt"
	"log"
	"net/http"
	"twitter-clone/internal/infrastructure/repository"
	httpHandler "twitter-clone/internal/interface/http"
	"twitter-clone/internal/usecase"
	"twitter-clone/pkg/middleware"
)

func main() {
	// Initialize database
	db := repository.NewInMemoryDatabase()

	// Initialize repositories
	userRepo := db.UserRepository()
	postRepo := db.PostRepository()
	hashtagRepo := db.HashtagRepository()

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(userRepo)
	postUseCase := usecase.NewPostUseCase(postRepo, userRepo, hashtagRepo)

	// Initialize handlers
	authHandler := httpHandler.NewAuthHandler(authUseCase)
	postHandler := httpHandler.NewPostHandler(postUseCase)

	// Setup routes
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/login", authHandler.Login)
	
	// Protected auth routes
	mux.HandleFunc("/api/user/profile", middleware.AuthMiddleware(authHandler.GetProfile))
	mux.HandleFunc("/api/user/follow", middleware.AuthMiddleware(authHandler.FollowUser))
	mux.HandleFunc("/api/user/unfollow", middleware.AuthMiddleware(authHandler.UnfollowUser))
	mux.HandleFunc("/api/user/upgrade", middleware.AuthMiddleware(authHandler.UpgradeSubscription))
	mux.HandleFunc("/api/user/tokens", middleware.AuthMiddleware(authHandler.AddTokens))
	mux.HandleFunc("/api/users/search", middleware.AuthMiddleware(authHandler.SearchUsers))

	// Post routes
	mux.HandleFunc("/api/posts", middleware.AuthMiddleware(postHandler.CreatePost))
	mux.HandleFunc("/api/posts/all", postHandler.GetAllPosts)
	mux.HandleFunc("/api/posts/popular", postHandler.GetPopularPosts)
	mux.HandleFunc("/api/posts/user", postHandler.GetUserPosts)
	mux.HandleFunc("/api/posts/get", postHandler.GetPost)
	mux.HandleFunc("/api/posts/like", middleware.AuthMiddleware(postHandler.LikePost))
	mux.HandleFunc("/api/posts/delete", middleware.AuthMiddleware(postHandler.DeletePost))
	mux.HandleFunc("/api/posts/edit", middleware.AuthMiddleware(postHandler.EditPost))
	mux.HandleFunc("/api/posts/reply", middleware.AuthMiddleware(postHandler.ReplyToPost))
	mux.HandleFunc("/api/posts/search", postHandler.SearchPosts)
	mux.HandleFunc("/api/posts/replies", postHandler.GetReplies)
	mux.HandleFunc("/api/posts/share", postHandler.GenerateShareLink)
	mux.HandleFunc("/api/posts/recommend", middleware.AuthMiddleware(postHandler.RecommendPosts))
	mux.HandleFunc("/api/posts/sort", postHandler.SortPosts)
	mux.HandleFunc("/api/posts/report", middleware.AuthMiddleware(postHandler.ReportPost))

	// CORS middleware
	handler := enableCORS(mux)

	// Start server
	port := ":8080"
	fmt.Printf("🚀 Twitter Clone API Server starting on port %s\n", port)
	fmt.Println("📝 Endpoints:")
	fmt.Println("   POST /api/register - Register new user")
	fmt.Println("   POST /api/login - Login user")
	fmt.Println("   POST /api/posts - Create post (auth required)")
	fmt.Println("   GET  /api/posts/popular - Get popular posts")
	fmt.Println("   GET  /api/posts/search?q=query - Search posts")
	fmt.Println("   POST /api/posts/like - Like/unlike post (auth required)")
	fmt.Println("   GET  /api/users/search?q=query - Search users (auth required)")
	fmt.Println("   POST /api/user/follow - Follow user (auth required)")
	fmt.Println("   POST /api/user/upgrade - Upgrade subscription (auth required)")
	
	log.Fatal(http.ListenAndServe(port, handler))
}

// enableCORS adds CORS headers to responses
func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		handler.ServeHTTP(w, r)
	})
}
