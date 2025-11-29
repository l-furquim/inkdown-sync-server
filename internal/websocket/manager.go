package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type ClientMessage struct {
	Client  *Client
	Message []byte
}

type Manager struct {
	clients        map[string]*Client
	userIndex      map[string]map[string]bool
	clientsMutex   sync.RWMutex
	Register       chan *Client
	Unregister     chan *Client
	HandleMessage  chan *ClientMessage
	maxConnPerUser int
	writeWait      time.Duration
	pongWait       time.Duration
	pingPeriod     time.Duration
	messageHandler MessageHandler
}

type MessageHandler interface {
	HandleWebSocketMessage(client *Client, msg *Message) error
}

func NewManager(maxConnPerUser int, writeWait, pongWait, pingPeriod time.Duration) *Manager {
	return &Manager{
		clients:        make(map[string]*Client),
		userIndex:      make(map[string]map[string]bool),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		HandleMessage:  make(chan *ClientMessage),
		maxConnPerUser: maxConnPerUser,
		writeWait:      writeWait,
		pongWait:       pongWait,
		pingPeriod:     pingPeriod,
	}
}

func (m *Manager) SetMessageHandler(handler MessageHandler) {
	m.messageHandler = handler
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.registerClient(client)

		case client := <-m.Unregister:
			m.unregisterClient(client)

		case clientMsg := <-m.HandleMessage:
			m.processMessage(clientMsg)
		}
	}
}

func (m *Manager) registerClient(client *Client) {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	if m.userIndex[client.UserID] == nil {
		m.userIndex[client.UserID] = make(map[string]bool)
	}

	if len(m.userIndex[client.UserID]) >= m.maxConnPerUser {
		log.Printf("max connections reached for user %s", client.UserID)
		close(client.Send)
		return
	}

	m.clients[client.ID] = client
	m.userIndex[client.UserID][client.ID] = true

	log.Printf("client registered: %s (user: %s, device: %s)", client.ID, client.UserID, client.DeviceID)
}

func (m *Manager) unregisterClient(client *Client) {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	if _, ok := m.clients[client.ID]; ok {
		delete(m.clients, client.ID)
		delete(m.userIndex[client.UserID], client.ID)

		if len(m.userIndex[client.UserID]) == 0 {
			delete(m.userIndex, client.UserID)
		}

		close(client.Send)
		log.Printf("client unregistered: %s", client.ID)
	}
}

func (m *Manager) processMessage(clientMsg *ClientMessage) {
	var msg Message
	if err := json.Unmarshal(clientMsg.Message, &msg); err != nil {
		log.Printf("error unmarshaling message: %v", err)
		return
	}

	if m.messageHandler != nil {
		if err := m.messageHandler.HandleWebSocketMessage(clientMsg.Client, &msg); err != nil {
			log.Printf("error handling message: %v", err)
		}
	}
}

func (m *Manager) BroadcastToUser(userID string, message *Message, excludeDeviceID string) error {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	clientIDs, exists := m.userIndex[userID]
	if !exists {
		return nil
	}

	for clientID := range clientIDs {
		client := m.clients[clientID]
		if client.DeviceID != excludeDeviceID {
			select {
			case client.Send <- messageBytes:
			default:
				log.Printf("client %s send buffer full, closing connection", clientID)
				m.Unregister <- client
			}
		}
	}

	return nil
}

func (m *Manager) SendToClient(clientID string, message *Message) error {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	client, exists := m.clients[clientID]
	if !exists {
		return nil
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case client.Send <- messageBytes:
	default:
		log.Printf("client %s send buffer full", clientID)
	}

	return nil
}

func (m *Manager) GetUserConnections(userID string) int {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	if clients, exists := m.userIndex[userID]; exists {
		return len(clients)
	}
	return 0
}
