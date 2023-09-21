package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserRoleKey  contextKey = "user_role"
	TenantIDKey  contextKey = "tenant_id"
	BranchIDKey  contextKey = "branch_id"
	RequestIDKey contextKey = "request_id"
)

// ExtractHeaders menyalin header identitas dari API Gateway ke context Go
func ExtractHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if userID := r.Header.Get("X-User-ID"); userID != "" {
				ctx = context.WithValue(ctx, UserIDKey, userID)
			}
			if role := r.Header.Get("X-User-Role"); role != "" {
				ctx = context.WithValue(ctx, UserRoleKey, role)
			}
			if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
				ctx = context.WithValue(ctx, TenantIDKey, tenantID)
			}
			if branchID := r.Header.Get("X-Branch-ID"); branchID != "" {
				ctx = context.WithValue(ctx, BranchIDKey, branchID)
			}
			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				ctx = context.WithValue(ctx, RequestIDKey, requestID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper getter functions
func GetUserID(ctx context.Context) string {
	val, _ := ctx.Value(UserIDKey).(string)
	return val
}

func GetUserRole(ctx context.Context) string {
	val, _ := ctx.Value(UserRoleKey).(string)
	return val
}

func GetTenantID(ctx context.Context) string {
	val, _ := ctx.Value(TenantIDKey).(string)
	return val
}

func GetBranchID(ctx context.Context) string {
	val, _ := ctx.Value(BranchIDKey).(string)
	return val
}

func GetRequestID(ctx context.Context) string {
	val, _ := ctx.Value(RequestIDKey).(string)
	return val
}
