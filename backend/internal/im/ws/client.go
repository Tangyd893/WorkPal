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
		return true // 开发环境允许所有 origin
	},
}

// Client 单个 WebSocket 客户端连接
type Client struct {
	UserID   int64                // 用户 ID
	Conn     *websocket.Conn      // WebSocket 连接
	Hub      *Hub                 // Hub 引用
	SendCh   chan []byte          // 发送队列
	done     chan struct{}
	doneOnce sync.Once
	mu       sync.Mutex
}

// NewClient 创建客户端
func NewClient(userID int64, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
		Hub:    hub,
		SendCh: make(chan []byte, 256),
		done:   make(chan struct{}),
	}
}

// Run 启动客户端读写循环
func (c *Client) Run(ctx context.Context) {
	// 注册到 Hub
	c.Hub.Register(c)

	// 启动读 goroutine
	go c.readLoop(ctx)

	// 启动写 goroutine
	go c.writeLoop()

	// 启动发送队列消费
	go c.sendLoop()
}

// readLoop 读取客户端消息
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

// writeLoop 从 SendCh 发送消息到客户端
func (c *Client) writeLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			// 发送 ping
			c.mu.Lock()
			if c.Conn != nil {
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.cleanup()
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

// sendLoop 消费发送队列
func (c *Client) sendLoop() {
	for {
		select {
		case <-c.done:
			return
		case msg := <-c.SendCh:
			// SendCh 有缓冲，写入即可
			if len(c.SendCh) == 0 {
				// channel 有空间，直接发
				select {
				case c.SendCh <- msg:
				default:
				}
			}
		}
	}
}

// handleMessage 处理收到的消息
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
	default:
		c.SendError("unknown message type: " + msg.Type)
	}
}

// handleChat 处理聊天消息
func (c *Client) handleChat(msg *WSMessage) {
	// msg.To 是目标（user_id 或 conv_id）
	// 根据 msg.ConvID 判断是否是会话 ID
	convID := msg.ConvID
	if convID <= 0 {
		// 私聊时 ConvID=0，To 字段是对方 user_id
		if _, err := strconv.ParseInt(msg.To, 10, 64); err == nil {
			// 私聊 - 需要业务层处理会话创建，Hub 只做路由
		}
	}

	// 将消息广播到会话房间
	if convID > 0 {
		wsData, _ := msg.Marshal()
		c.Hub.BroadcastToRoom(convID, c.UserID, wsData, c)
		// 发送 ACK
		ack := NewAckMsg(msg.Seq)
		ackData, _ := ack.Marshal()
		c.Send(ackData)
	}
}

// handlePing 处理心跳
func (c *Client) handlePing() {
	pong := &WSMessage{Type: TypePong}
	data, _ := pong.Marshal()
	c.Send(data)
}

// Send 发送消息到客户端
func (c *Client) Send(data []byte) {
	select {
	case c.SendCh <- data:
	default:
		// channel 满了，丢弃
		log.Printf("[WS] 发送队列已满 userID=%d", c.UserID)
	}
}

// SendError 发送错误消息
func (c *Client) SendError(text string) {
	err := NewErrorMsg(text)
	data, _ := err.Marshal()
	c.Send(data)
}

// Close 关闭连接
func (c *Client) Close() {
	c.cleanup()
}

// cleanup 清理资源
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

// GetHub 获取全局 Hub 实例（通过包级别变量）
var defaultHub *Hub
var hubOnce sync.Once
var hubMu sync.Mutex

// InitHub 初始化全局 Hub
func InitHub() *Hub {
	hubMu.Lock()
	defer hubMu.Unlock()
	hubOnce.Do(func() {
		defaultHub = NewHub()
		go defaultHub.Run()
	})
	return defaultHub
}

// GetHub 获取全局 Hub
func GetHub() *Hub {
	hubMu.Lock()
	defer hubMu.Unlock()
	return defaultHub
}

// UpgradeToWebSocket 将 HTTP 连接升级为 WebSocket
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

// ParseTokenFromQuery 从 URL query 解析 token（简化版，实际应在 handler 层 JWT 验证后传入 userID）
func ParseTokenFromQuery(u string) (int64, error) {
	// 格式: /ws?token=xxx 或 /ws?user_id=123
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
