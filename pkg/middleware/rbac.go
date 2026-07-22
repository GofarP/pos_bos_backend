package middleware

import (
	"fmt"
	"net/http"
	"pos_bos/internal/core/domain"
)

// RequirePermission checks if the user has a specific permission
func RequirePermission(permissionName string, rbacRepo domain.RBACRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDVal := r.Context().Value(UserIDKey)
			if userIDVal == nil {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// JWT unmarshals numbers as float64 by default
			var userID int
			switch v := userIDVal.(type) {
			case float64:
				userID = int(v)
			case int:
				userID = v
			default:
				http.Error(w, `{"error": "Invalid user ID format"}`, http.StatusInternalServerError)
				return
			}

			hasPermission, err := rbacRepo.HasPermission(r.Context(), userID, permissionName)
			if err != nil {
				// Avoid exposing internal DB errors directly in production, but log them
				fmt.Printf("Error checking permissions: %v\n", err)
				http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				http.Error(w, `{"error": "Forbidden: Insufficient Permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole checks if the user has a specific role
func RequireRole(roleName string, rbacRepo domain.RBACRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDVal := r.Context().Value(UserIDKey)
			if userIDVal == nil {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			var userID int
			switch v := userIDVal.(type) {
			case float64:
				userID = int(v)
			case int:
				userID = v
			default:
				http.Error(w, `{"error": "Invalid user ID format"}`, http.StatusInternalServerError)
				return
			}

			hasRole, err := rbacRepo.HasRole(r.Context(), userID, roleName)
			if err != nil {
				fmt.Printf("Error checking roles: %v\n", err)
				http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
				return
			}

			if !hasRole {
				http.Error(w, `{"error": "Forbidden: Insufficient Role"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
