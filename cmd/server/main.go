package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inkdown-sync-server/internal/config"
	"inkdown-sync-server/internal/handler"
	"inkdown-sync-server/internal/middleware"
	"inkdown-sync-server/internal/repository"
	"inkdown-sync-server/internal/service"

	_ "github.com/go-kivik/kivik/v4/couchdb"

	"github.com/go-kivik/kivik/v4"
	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	couchURL := fmt.Sprintf("http://%s:%s@%s:%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
	)

	client, err := kivik.New("couch", couchURL)
	if err != nil {
		log.Fatalf("Failed to connect to CouchDB: %v", err)
	}

	exists, err := client.DBExists(context.Background(), cfg.Database.Name)
	if err != nil {
		log.Fatalf("Failed to check database existence: %v", err)
	}

	if !exists {
		if err := client.CreateDB(context.Background(), cfg.Database.Name); err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		log.Printf("Created database: %s", cfg.Database.Name)
	}

	userRepo := repository.NewUserRepository(client, cfg.Database.Name)

	authService := service.NewAuthService(
		userRepo,
		cfg.JWT.Secret,
		cfg.JWT.Expiration,
		cfg.JWT.RefreshTokenExpiration,
	)
	userService := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	router := mux.NewRouter()

	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	))

	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	apiRouter.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	apiRouter.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	apiRouter.HandleFunc("/auth/refresh", authHandler.Refresh).Methods("POST")

	protected := apiRouter.NewRoute().Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	protected.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")
	protected.HandleFunc("/users/me", userHandler.GetMe).Methods("GET")
	protected.HandleFunc("/users/me", userHandler.UpdateMe).Methods("PUT")

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting Inkdown Sync Server on %s (env: %s)", addr, cfg.Server.Env)
		log.Printf("Connected to CouchDB at %s:%s", cfg.Database.Host, cfg.Database.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"inkdown-sync-server"}`))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Inkdown Sync Server API","version":"1.0.0","endpoints":{"/api/v1/auth/register":"POST","/api/v1/auth/login":"POST","/api/v1/auth/refresh":"POST","/api/v1/users/me":"GET (protected)"}}`))
}
