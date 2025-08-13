package middleware

import (
	"Effective_Mobile/internal/models"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"net/http"
)

// RateLimiterMiddleware creates a new rate limiting middleware.
// r is the maximum number of events per second.
// b is the maximum burst size, the maximum number of events that can happen at a single moment.
func RateLimiterMiddleware(r rate.Limit, b int, log *zap.Logger) func(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(r, b)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !limiter.Allow() {
				log.Warn("Rate limit exceeded", zap.String("ip", req.RemoteAddr))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				response := models.Response{
					Status: http.StatusTooManyRequests,
					Msg:    "Too many requests",
				}
				json.NewEncoder(w).Encode(response)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

func LoggingMiddleware(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestLog := log.With(
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("remote_addr", req.RemoteAddr),
			)

			requestLog.Info("Request started")
			ctx := context.WithValue(req.Context(), "logger", requestLog)
			next.ServeHTTP(w, req.WithContext(ctx))
			requestLog.Info("Request completed")
		})
	}
}
