package middleware

import (
	"context"
	"net/http"
	"strings"
	"twitter-clone/pkg/token"
)

type contextKey string

const userIDKey contextKey = "userID"
const usernameKey contextKey = "username"
const userBadgeKey contextKey = "userBadge"

// AuthMiddleware validates JWT tokens
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims, err := token.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, usernameKey, claims.Username)
		ctx = context.WithValue(ctx, userBadgeKey, claims.Badge)

		next(w, r.WithContext(ctx))
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(r *http.Request) uint64 {
	userID, ok := r.Context().Value(userIDKey).(uint64)
	if !ok {
		return 0
	}
	return userID
}

// GetUsernameFromContext extracts username from context
func GetUsernameFromContext(r *http.Request) string {
	username, ok := r.Context().Value(usernameKey).(string)
	if !ok {
		return ""
	}
	return username
}

// GetUserBadgeFromContext extracts user badge from context
func GetUserBadgeFromContext(r *http.Request) string {
	badge, ok := r.Context().Value(userBadgeKey).(string)
	if !ok {
		return ""
	}
	return badge
}

// AdminMiddleware checks if user is admin
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In production, check admin rights properly
		// For now, just pass through
		next(w, r)
	}
}
