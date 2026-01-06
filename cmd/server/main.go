package main

import (
	"context"
	"encoding/json"
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
	"inkdown-sync-server/internal/websocket"

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
	deviceRepo := repository.NewDeviceRepository(client, cfg.Database.Name)
	keyStoreRepo := repository.NewKeyStoreRepository(client, cfg.Database.Name)
	noteRepo := repository.NewNoteRepository(client, cfg.Database.Name)
	workspaceRepo := repository.NewWorkspaceRepository(client, cfg.Database.Name)
	cliTokenRepo := repository.NewCLITokenRepository(client, cfg.Database.Name)

	baseURL := fmt.Sprintf("%s/%s", couchURL, cfg.Database.Name)
	versionRepo := repository.NewNoteVersionRepository(baseURL)
	syncMetadataRepo := repository.NewSyncMetadataRepository(baseURL)
	conflictRepo := repository.NewConflictRepository(baseURL)

	// WebSocket Manager
	wsManager := websocket.NewManager(
		cfg.WebSocket.MaxConnPerUser,
		cfg.WebSocket.WriteWait,
		cfg.WebSocket.PongWait,
		cfg.WebSocket.PingPeriod,
	)
	go wsManager.Run()

	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiration, cfg.JWT.RefreshTokenExpiration)
	userService := service.NewUserService(userRepo)
	deviceService := service.NewDeviceService(deviceRepo)
	securityService := service.NewSecurityService(keyStoreRepo)
	cliTokenService := service.NewCLITokenService(cliTokenRepo, userRepo)

	syncService := service.NewSyncService(noteRepo, versionRepo, syncMetadataRepo, wsManager)
	conflictService := service.NewConflictService(conflictRepo, versionRepo, noteRepo)
	noteService := service.NewNoteService(noteRepo, versionRepo, conflictService, syncService)
	workspaceService := service.NewWorkspaceService(workspaceRepo, noteRepo)

	wsMessageHandler := handler.NewWebSocketMessageHandler(syncService)
	wsManager.SetMessageHandler(wsMessageHandler)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	securityHandler := handler.NewSecurityHandler(securityService)
	noteHandler := handler.NewNoteHandler(noteService)
	wsHandler := handler.NewWebSocketHandler(wsManager, cfg.JWT.Secret)
	syncHandler := handler.NewSyncHandler(syncService, conflictService)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
	cliTokenHandler := handler.NewCLITokenHandler(cliTokenService)

	r := mux.NewRouter()

	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	))

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/refresh", authHandler.Refresh).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST", "OPTIONS")

	api.HandleFunc("/cli/login", cliTokenHandler.Login).Methods("POST", "OPTIONS")
	api.HandleFunc("/cli/validate", cliTokenHandler.Validate).Methods("POST", "OPTIONS")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))

	protected.HandleFunc("/users/me", userHandler.GetMe).Methods("GET", "OPTIONS")
	protected.HandleFunc("/users/me", userHandler.UpdateMe).Methods("PUT", "OPTIONS")

	protected.HandleFunc("/cli/tokens", cliTokenHandler.Create).Methods("POST", "OPTIONS")
	protected.HandleFunc("/cli/tokens", cliTokenHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/cli/tokens/{id}", cliTokenHandler.Get).Methods("GET", "OPTIONS")
	protected.HandleFunc("/cli/tokens/{id}/revoke", cliTokenHandler.Revoke).Methods("POST", "OPTIONS")
	protected.HandleFunc("/cli/tokens/{id}", cliTokenHandler.Delete).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/devices", deviceHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/devices/register", deviceHandler.Register).Methods("POST", "OPTIONS")
	protected.HandleFunc("/devices/{id}", deviceHandler.Revoke).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/security/keys/setup", securityHandler.UploadKey).Methods("POST", "OPTIONS")
	protected.HandleFunc("/security/keys/sync", securityHandler.GetKey).Methods("GET", "OPTIONS")

	protected.HandleFunc("/notes", noteHandler.Create).Methods("POST", "OPTIONS")
	protected.HandleFunc("/notes", noteHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/notes/{id}", noteHandler.Get).Methods("GET", "OPTIONS")
	protected.HandleFunc("/notes/{id}", noteHandler.Update).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/notes/{id}", noteHandler.Delete).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/workspaces", workspaceHandler.Create).Methods("POST", "OPTIONS")
	protected.HandleFunc("/workspaces", workspaceHandler.List).Methods("GET", "OPTIONS")
	protected.HandleFunc("/workspaces/{id}", workspaceHandler.Get).Methods("GET", "OPTIONS")
	protected.HandleFunc("/workspaces/{id}", workspaceHandler.Update).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/workspaces/{id}", workspaceHandler.Delete).Methods("DELETE", "OPTIONS")

	protected.HandleFunc("/sync/request", syncHandler.ProcessSync).Methods("POST", "OPTIONS")
	protected.HandleFunc("/sync/changes", syncHandler.GetChanges).Methods("GET", "OPTIONS")
	protected.HandleFunc("/sync/manifest", syncHandler.GetManifest).Methods("GET", "OPTIONS")
	protected.HandleFunc("/sync/batch-diff", syncHandler.BatchDiff).Methods("POST", "OPTIONS")
	protected.HandleFunc("/sync/conflicts", syncHandler.ListConflicts).Methods("GET", "OPTIONS")
	protected.HandleFunc("/sync/resolve/{id}", syncHandler.ResolveConflict).Methods("POST", "OPTIONS")

	// These routes use CLI tokens (ink_xxxxx) instead of JWT
	cliProtected := api.PathPrefix("/community").Subrouter()
	cliProtected.Use(middleware.CLIAuthMiddleware(cliTokenService))
	cliProtected.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(string)
		scopes := r.Context().Value("cli_scopes").([]string)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "CLI authentication successful",
			"user_id": userID,
			"scopes":  scopes,
		})
	}).Methods("GET", "OPTIONS")

	r.HandleFunc("/ws", wsHandler.HandleConnection)

	// Health endpoint
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
