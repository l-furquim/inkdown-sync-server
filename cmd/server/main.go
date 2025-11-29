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
	// Repositories
	userRepo := repository.NewUserRepository(client, cfg.Database.Name)
	deviceRepo := repository.NewDeviceRepository(client, cfg.Database.Name)
	keyStoreRepo := repository.NewKeyStoreRepository(client, cfg.Database.Name)
	noteRepo := repository.NewNoteRepository(client, cfg.Database.Name)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiration, cfg.JWT.RefreshTokenExpiration)
	userService := service.NewUserService(userRepo)
	deviceService := service.NewDeviceService(deviceRepo)
	securityService := service.NewSecurityService(keyStoreRepo)
	noteService := service.NewNoteService(noteRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	securityHandler := handler.NewSecurityHandler(securityService)
	noteHandler := handler.NewNoteHandler(noteService)

	// Router
	r := mux.NewRouter()

	// Middlewares
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	))

	// API Router
	api := r.PathPrefix("/api/v1").Subrouter()

	// Public Routes
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/refresh", authHandler.Refresh).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST", "OPTIONS")

	// Protected Routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))

	// User Routes
	protected.HandleFunc("/users/me", userHandler.GetMe).Methods("GET", "OPTIONS")
	protected.HandleFunc("/users/me", userHandler.UpdateMe).Methods("PUT", "OPTIONS")

	// Device Routes
	protected.HandleFunc("/devices", deviceHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/devices/register", deviceHandler.Register).Methods("POST", "OPTIONS")
	protected.HandleFunc("/devices/{id}", deviceHandler.Revoke).Methods("DELETE", "OPTIONS")

	// Security Routes (E2EE)
	protected.HandleFunc("/security/keys/setup", securityHandler.UploadKey).Methods("POST", "OPTIONS")
	protected.HandleFunc("/security/keys/sync", securityHandler.GetKey).Methods("GET", "OPTIONS")

	// Note Routes
	protected.HandleFunc("/notes", noteHandler.Create).Methods("POST", "OPTIONS")
	protected.HandleFunc("/notes", noteHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/notes/{id}", noteHandler.Get).Methods("GET", "OPTIONS")
	protected.HandleFunc("/notes/{id}", noteHandler.Update).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/notes/{id}", noteHandler.Delete).Methods("DELETE", "OPTIONS")

	// Health and Root handlers (public)
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/", rootHandler).Methods("GET")

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
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
