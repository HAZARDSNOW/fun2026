package main

import (
	"auth-service/internal/infrastructure/repository"
	authhttp "auth-service/internal/interface/http"
	"auth-service/internal/usecase"
	"auth-service/pkg/hash"
	"auth-service/pkg/token"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Infrastructure layer
	userRepo := repository.NewInMemoryUserRepository()
	hasher := hash.NewBcryptHasher(10)
	jwtManager := token.NewJWTManager("your-secret-key-change-in-production", 24*time.Hour)

	// Use case layer
	authUsecase := usecase.NewAuthUsecase(userRepo, hasher, jwtManager)

	// Interface layer
	authHandler := authhttp.NewAuthHandler(authUsecase)

	// Routes
	http.HandleFunc("/api/register", authHandler.Register)
	http.HandleFunc("/api/login", authHandler.Login)

	// Start server
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
