package ws

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	UserID   int64
	Conn     *websocket.Conn
	Hub      *Hub
	SendCh   chan []byte
	done     chan struct{}
	doneOnce sync.Once
	mu       sync.Mutex
}

func NewClient(userID int64, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
		Hub:    hub,
		SendCh: make(chan []byte, 256),
		done:   make(chan struct{}),
	}
}

func (c *Client) Run(ctx context.Context) {
	c.Hub.Register(c)
	go c.readLoop(ctx)
	go c.writeLoop()
	go c.sendLoop()
}

func (c *Client) readLoop(ctx context.Context) {
	defer func() {
		c.cleanup()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
			_, data, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[WS] 读取错误 userID=%d: %v", c.UserID, err)
				}
				return
			}
			c.handleMessage(data)
		}
	}
}

func (c *Client) writeLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.Conn != nil {
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.cleanup()
					c.mu.Unlock()
					return
				}
			}
			c.mu.Unlock()
		case msg, ok := <-c.SendCh:
			if !ok {
				return
			}
			c.mu.Lock()
			if c.Conn != nil {
				if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					c.cleanup()
					c.mu.Unlock()
					return
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) sendLoop() {
	for {
		select {
		case <-c.done:
			return
		case msg := <-c.SendCh:
			if len(c.SendCh) == 0 {
				select {
				case c.SendCh <- msg:
				default:
				}
			}
		}
	}
}

func (c *Client) handleMessage(data []byte) {
	msg, err := Unmarshal(data)
	if err != nil {
		c.SendError("invalid message format")
		return
	}

	switch msg.Type {
	case TypeChat:
		c.handleChat(msg)
	case TypePing:
		c.handlePing()
	case TypeRead:
		c.handleRead(msg)
	case TypeReadAll:
		c.handleReadAll(msg)
	default:
		c.SendError("unknown message type: " + msg.Type)
	}
}

func (c *Client) handleChat(msg *WSMessage) {
	convID := msg.ConvID
	// 私聊时 ConvID=0，To 字段是对方 user_id：
	// TODO: 私聊路由需业务层处理，Hub 这里只做占位

	if convID > 0 {
		wsData, _ := msg.Marshal()
		c.Hub.BroadcastToRoom(convID, c.UserID, wsData, c)
		ack := NewAckMsg(msg.Seq)
		ackData, _ := ack.Marshal()
		c.Send(ackData)
	}
}

func (c *Client) handlePing() {
	pong := &WSMessage{Type: TypePong}
	data, _ := pong.Marshal()
	c.Send(data)
}

// handleRead 处理已读回执（用户阅读消息后主动发送）
func (c *Client) handleRead(msg *WSMessage) {
	if msg.ConvID <= 0 {
		return
	}
	readData, _ := msg.Marshal()
	// 广播给房间内所有其他成员，告知谁已读了哪条消息
	c.Hub.BroadcastToRoom(msg.ConvID, c.UserID, readData, c)
}

// handleReadAll 处理全部已读（用户打开会话时告知全部已读）
func (c *Client) handleReadAll(msg *WSMessage) {
	if msg.ConvID <= 0 {
		return
	}
	readAllData, _ := msg.Marshal()
	c.Hub.BroadcastToRoom(msg.ConvID, c.UserID, readAllData, c)
}

func (c *Client) Send(data []byte) {
	select {
	case c.SendCh <- data:
	default:
		log.Printf("[WS] 发送队列已满 userID=%d", c.UserID)
	}
}

func (c *Client) SendError(text string) {
	err := NewErrorMsg(text)
	data, _ := err.Marshal()
	c.Send(data)
}

func (c *Client) Close() {
	c.cleanup()
}

func (c *Client) cleanup() {
	c.doneOnce.Do(func() {
		close(c.done)
		c.Hub.Unregister(c)
		c.mu.Lock()
		if c.Conn != nil {
			c.Conn.Close()
			c.Conn = nil
		}
		c.mu.Unlock()
		close(c.SendCh)
	})
}

var defaultHub *Hub
var hubOnce sync.Once
var hubMu sync.Mutex

func InitHub() *Hub {
	hubMu.Lock()
	defer hubMu.Unlock()
	hubOnce.Do(func() {
		defaultHub = NewHub()
		go defaultHub.Run()
	})
	return defaultHub
}

func GetHub() *Hub {
	hubMu.Lock()
	defer hubMu.Unlock()
	return defaultHub
}

func UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, userID int64) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	hub := GetHub()
	if hub == nil {
		conn.Close()
		return nil, ErrHubNotInit
	}
	client := NewClient(userID, conn, hub)
	client.Run(context.Background())
	return conn, nil
}

func ParseTokenFromQuery(u string) (int64, error) {
	if strings.HasPrefix(u, "/ws?") {
		params := strings.TrimPrefix(u, "/ws?")
		pairs := strings.Split(params, "&")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 && kv[0] == "user_id" {
				uid, err := strconv.ParseInt(kv[1], 10, 64)
				if err == nil {
					return uid, nil
				}
			}
		}
	}
	return 0, nil
}

var ErrHubNotInit = &AppError{Code: 50000, Message: "WebSocket Hub 未初始化"}

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}
