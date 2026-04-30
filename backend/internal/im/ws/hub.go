package ws

import (
	"context"
	"log"
	"sync"
)

type Hub struct {
	clients    map[int64]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMsg

	rooms  map[int64]map[*Client]bool
	roomMu sync.RWMutex

	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	cluster clusterRoomBroadcaster
}

type clusterRoomBroadcaster interface {
	BroadcastRoom(ctx context.Context, convID int64, fromID int64, content []byte) error
}

type BroadcastMsg struct {
	ConvID  int64
	FromID  int64
	Content []byte
	Exclude *Client
}

func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:    make(map[int64]*Client),
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		broadcast:  make(chan *BroadcastMsg, 100),
		rooms:      make(map[int64]map[*Client]bool),
		running:    true,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			count := len(h.clients)
			h.mu.Unlock()
			log.Printf("[hub] user %d connected (total=%d)", client.UserID, count)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				count := len(h.clients)

				h.roomMu.Lock()
				for convID, members := range h.rooms {
					if _, exists := members[client]; exists {
						delete(members, client)
						log.Printf("[hub] user %d left room %d", client.UserID, convID)
					}
				}
				h.roomMu.Unlock()
				log.Printf("[hub] user %d disconnected (total=%d)", client.UserID, count)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.deliverMessage(msg)
		}
	}
}

func (h *Hub) Stop() {
	h.cancel()
	h.mu.Lock()
	h.running = false
	h.mu.Unlock()
}

func (h *Hub) SetClusterBroadcaster(cluster clusterRoomBroadcaster) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cluster = cluster
}

func (h *Hub) Register(client *Client) {
	if !h.isRunning() {
		return
	}
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	if !h.isRunning() {
		return
	}
	h.unregister <- client
}

func (h *Hub) JoinRoom(client *Client, convID int64) {
	h.roomMu.Lock()
	defer h.roomMu.Unlock()
	if _, ok := h.rooms[convID]; !ok {
		h.rooms[convID] = make(map[*Client]bool)
	}
	h.rooms[convID][client] = true
	log.Printf("[hub] user %d joined room %d", client.UserID, convID)
}

func (h *Hub) LeaveRoom(client *Client, convID int64) {
	h.roomMu.Lock()
	defer h.roomMu.Unlock()
	if members, ok := h.rooms[convID]; ok {
		if _, exists := members[client]; exists {
			delete(members, client)
			log.Printf("[hub] user %d left room %d", client.UserID, convID)
		}
	}
}

func (h *Hub) BroadcastToRoom(convID int64, fromID int64, content []byte, exclude *Client) {
	msg := &BroadcastMsg{
		ConvID:  convID,
		FromID:  fromID,
		Content: content,
		Exclude: exclude,
	}
	select {
	case <-h.ctx.Done():
		return
	case h.broadcast <- msg:
	default:
		log.Printf("[hub] local broadcast queue full convID=%d", convID)
	}

	h.mu.RLock()
	cluster := h.cluster
	ctx := h.ctx
	h.mu.RUnlock()
	if cluster != nil {
		if err := cluster.BroadcastRoom(ctx, convID, fromID, content); err != nil {
			log.Printf("[hub] cluster room broadcast failed convID=%d: %v", convID, err)
		}
	}
}

func (h *Hub) SendToUser(userID int64, content []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if client, ok := h.clients[userID]; ok {
		client.Send(content)
	}
}

func (h *Hub) SendToUsers(userIDs []int64, content []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, userID := range userIDs {
		if client, ok := h.clients[userID]; ok {
			client.Send(content)
		}
	}
}

func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

func (h *Hub) deliverMessage(msg *BroadcastMsg) {
	h.roomMu.RLock()
	defer h.roomMu.RUnlock()
	if members, ok := h.rooms[msg.ConvID]; ok {
		for client := range members {
			if msg.Exclude != nil && client == msg.Exclude {
				continue
			}
			client.Send(msg.Content)
		}
	}
}

func (h *Hub) isRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}
