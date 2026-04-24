package ws

import (
	"context"
	"log"
	"sync"
)

// Hub WebSocket 连接管理中心
type Hub struct {
	// 客户端管理
	clients    map[int64]*Client       // userID -> Client
	register   chan *Client            // 注册请求
	unregister chan *Client            // 注销请求
	broadcast  chan *BroadcastMsg     // 广播消息

	// 房间（会话）管理
	rooms      map[int64]map[*Client]bool // convID -> 客户端集合
	roomMu     sync.RWMutex

	// Hub 状态
	mu         sync.RWMutex
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
}

// BroadcastMsg 广播消息结构
type BroadcastMsg struct {
	ConvID  int64
	FromID  int64   // 发送者 userID
	Content []byte  // WSMessage JSON
	Exclude *Client // 排除的客户端（不发给自己）
}

// NewHub 创建 Hub
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

// Run 启动 Hub（阻塞）
func (h *Hub) Run() {
	for {
		select {
		case <-h.ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("[Hub] 用户 %d 已连接 (共 %d 连接)", client.UserID, len(h.clients))
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				// 从所有房间移除
				h.roomMu.Lock()
				for convID, members := range h.rooms {
					if _, exists := members[client]; exists {
						delete(members, client)
						log.Printf("[Hub] 用户 %d 离开房间 %d", client.UserID, convID)
					}
				}
				h.roomMu.Unlock()
				log.Printf("[Hub] 用户 %d 已断开 (共 %d 连接)", client.UserID, len(h.clients))
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.deliverMessage(msg)
		}
	}
}

// Stop 停止 Hub
func (h *Hub) Stop() {
	h.cancel()
	h.mu.Lock()
	h.running = false
	h.mu.Unlock()
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	if !h.isRunning() {
		return
	}
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	if !h.isRunning() {
		return
	}
	h.unregister <- client
}

// JoinRoom 加入房间（会话）
func (h *Hub) JoinRoom(client *Client, convID int64) {
	h.roomMu.Lock()
	defer h.roomMu.Unlock()
	if _, ok := h.rooms[convID]; !ok {
		h.rooms[convID] = make(map[*Client]bool)
	}
	h.rooms[convID][client] = true
	log.Printf("[Hub] 用户 %d 加入房间 %d", client.UserID, convID)
}

// LeaveRoom 离开房间
func (h *Hub) LeaveRoom(client *Client, convID int64) {
	h.roomMu.Lock()
	defer h.roomMu.Unlock()
	if members, ok := h.rooms[convID]; ok {
		if _, exists := members[client]; exists {
			delete(members, client)
			log.Printf("[Hub] 用户 %d 离开房间 %d", client.UserID, convID)
		}
	}
}

// BroadcastToRoom 向房间广播消息
func (h *Hub) BroadcastToRoom(convID int64, fromID int64, content []byte, exclude *Client) {
	h.broadcast <- &BroadcastMsg{
		ConvID:  convID,
		FromID:  fromID,
		Content: content,
		Exclude: exclude,
	}
}

// SendToUser 向指定用户发送消息（如果在线）
func (h *Hub) SendToUser(userID int64, content []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if client, ok := h.clients[userID]; ok {
		client.Send(content)
	}
}

// SendToUsers 向多个用户发送消息
func (h *Hub) SendToUsers(userIDs []int64, content []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, uid := range userIDs {
		if client, ok := h.clients[uid]; ok {
			client.Send(content)
		}
	}
}

// GetOnlineCount 获取在线人数
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// deliverMessage 投递消息到房间
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
