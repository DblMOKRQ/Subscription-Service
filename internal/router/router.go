package router

import (
	"Effective_Mobile/internal/middleware"
	"Effective_Mobile/internal/router/handlers"
	"context"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "Effective_Mobile/docs" // docs генерируется Swag
	"github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type Router struct {
	mux         *http.ServeMux
	log         *zap.Logger
	subsHandler *handlers.SubscriptionHandler
	server      *http.Server
}

func NewRouter(subsHandler *handlers.SubscriptionHandler, log *zap.Logger) *Router {
	return &Router{
		mux:         http.NewServeMux(),
		log:         log.Named("request"),
		subsHandler: subsHandler,
	}
}

func (r *Router) RunRouter(addr string, requestPerSec int, burst int) error {
	// Apply rate limiting middleware to all routes
	// 1 request per second, with a burst of 5 requests
	rateLimitedMux := middleware.RateLimiterMiddleware(rate.Limit(requestPerSec), burst, r.log)(r.mux)
	loggingMux := middleware.LoggingMiddleware(r.log)

	// Настройка обработчиков
	r.mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // URL указывает на сгенерированный файл
	))
	r.mux.HandleFunc("POST /subscriptions", r.subsHandler.CreateSubs)
	r.mux.HandleFunc("GET /subscriptions", r.subsHandler.GetSubs)
	r.mux.HandleFunc("PUT /subscriptions", r.subsHandler.UpdateSubs)
	r.mux.HandleFunc("DELETE /subscriptions", r.subsHandler.DeleteSubs)
	r.mux.HandleFunc("POST /subscriptions/summary", r.subsHandler.GetSummary)
	r.mux.HandleFunc("GET /all-subscriptions", r.subsHandler.ListSubs)

	r.server = &http.Server{
		Addr:    addr,
		Handler: loggingMux(rateLimitedMux),
	}

	serverErr := make(chan error, 1)

	go func() {
		r.log.Info("Starting server", zap.String("addr", addr))
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		r.log.Info("Received signal", zap.String("signal", sig.String()))
	case err := <-serverErr:
		r.log.Error("Server error", zap.Error(err))
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r.log.Info("Shutting down server...")
	if err := r.server.Shutdown(ctx); err != nil {
		r.log.Error("Forced shutdown", zap.Error(err))
		return err
	}

	r.log.Info("Server stopped gracefully")
	return nil
}

//func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
//		requestLog := r.log.With(
//			zap.String("method", req.Method),
//			zap.String("path", req.URL.Path),
//			zap.String("remote_addr", req.RemoteAddr),
//		)
//
//		requestLog.Info("Request started")
//		ctx := context.WithValue(req.Context(), "logger", requestLog)
//		next.ServeHTTP(w, req.WithContext(ctx))
//		requestLog.Info("Request completed")
//	})
//}
