package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"inkdown-sync-server/internal/domain"
	"inkdown-sync-server/internal/service"
	"inkdown-sync-server/internal/websocket"
	"inkdown-sync-server/pkg/jwt"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	manager   *websocket.Manager
	jwtSecret string
	upgrader  ws.Upgrader
}

func NewWebSocketHandler(manager *websocket.Manager, jwtSecret string) *WebSocketHandler {
	return &WebSocketHandler{
		manager:   manager,
		jwtSecret: jwtSecret,
		upgrader: ws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		log.Printf("[WebSocket] Missing authorization token")
		http.Error(w, "missing authorization token", http.StatusUnauthorized)
		return
	}

	log.Printf("[WebSocket] Validating token for connection")
	claims, err := jwt.ValidateToken(token, h.jwtSecret)
	if err != nil {
		log.Printf("[WebSocket] Token validation failed: %v", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		deviceID = "default"
	}

	log.Printf("[WebSocket] Upgrading connection for user: %s", userID)
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Failed to upgrade connection: %v", err)
		return
	}

	log.Printf("[WebSocket] Connection upgraded successfully for user: %s", userID)

	clientID := uuid.New().String()
	client := websocket.NewClient(clientID, userID, deviceID, conn, h.manager)

	h.manager.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

type WebSocketMessageHandler struct {
	syncService *service.SyncService
}

func NewWebSocketMessageHandler(syncService *service.SyncService) *WebSocketMessageHandler {
	return &WebSocketMessageHandler{
		syncService: syncService,
	}
}

func (h *WebSocketMessageHandler) HandleWebSocketMessage(client *websocket.Client, msg *websocket.Message) error {
	switch msg.Type {
	case websocket.TypeSyncRequest:
		return h.handleSyncRequest(client, msg)

	case websocket.TypePing:
		return h.handlePing(client)

	default:
		log.Printf("unknown message type: %s", msg.Type)
	}

	return nil
}

func (h *WebSocketMessageHandler) handleSyncRequest(client *websocket.Client, msg *websocket.Message) error {
	var payload websocket.SyncRequestPayload
	if err := msg.UnmarshalPayload(&payload); err != nil {
		return err
	}

	syncReq := &domain.SyncRequest{
		DeviceID:     payload.DeviceID,
		LastSyncTime: payload.LastSyncTime,
		NoteVersions: payload.NoteVersions,
	}

	response, err := h.syncService.ProcessSyncRequest(client.UserID, payload.DeviceID, syncReq)
	if err != nil {
		return err
	}

	responseMsg, err := websocket.NewMessage(websocket.TypeSyncResponse, &websocket.SyncResponsePayload{
		Changes:  convertToWSChanges(response.Changes),
		HasMore:  response.HasMore,
		SyncTime: response.SyncTime,
	})
	if err != nil {
		return err
	}

	responseBytes, _ := json.Marshal(responseMsg)
	client.Send <- responseBytes

	return nil
}

func (h *WebSocketMessageHandler) handlePing(client *websocket.Client) error {
	pongMsg, err := websocket.NewMessage(websocket.TypePong, nil)
	if err != nil {
		return err
	}

	pongBytes, _ := json.Marshal(pongMsg)
	client.Send <- pongBytes

	return nil
}

func convertToWSChanges(changes []*domain.NoteChange) []websocket.NoteChange {
	wsChanges := make([]websocket.NoteChange, len(changes))
	for i, c := range changes {
		data, _ := json.Marshal(c.Note)
		wsChanges[i] = websocket.NoteChange{
			NoteID:    c.NoteID,
			Operation: c.Operation,
			Version:   c.Version,
			Data:      data,
		}
	}
	return wsChanges
}
