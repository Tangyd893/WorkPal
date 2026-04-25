package ws

import (
	"sync"
	"testing"
)

// TestHubConcurrentAccess races all Hub operations together.
// Since we can't create real websocket.Conn, we test hub data structure directly.
func TestHubConcurrentAccess(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	N := 100
	var wg sync.WaitGroup

	// Simulate clients map access (the main shared state)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			uid := int64(i)

			// Test IsUserOnline / GetOnlineCount (read operations)
			hub.mu.RLock()
			_ = hub.clients[uid]
			hub.mu.RUnlock()

			_ = hub.IsUserOnline(uid)
			_ = hub.GetOnlineCount()
		}(i)
	}

	// Test rooms access (roomMu protected)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			convID := int64(i % 10)

			hub.roomMu.Lock()
			if _, ok := hub.rooms[convID]; !ok {
				hub.rooms[convID] = make(map[*Client]bool)
			}
			hub.roomMu.Unlock()

			// Broadcast (non-blocking channel send)
			hub.BroadcastToRoom(convID, 1, []byte("msg"), nil)
		}(i)
	}

	// Test broadcast concurrently
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			hub.BroadcastToRoom(int64(i%10), int64(i), []byte("test"), nil)
		}(i)
	}

	wg.Wait()
}

// TestHubRegisterUnregisterRace races Register vs Unregister
func TestHubRegisterUnregisterRace(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	clients := make([]*Client, 50)
	for i := 0; i < 50; i++ {
		clients[i] = &Client{UserID: int64(i), Hub: hub, SendCh: make(chan []byte, 256)}
	}

	var wg sync.WaitGroup

	// All register
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			hub.Register(c)
		}(clients[i])
	}
	wg.Wait()

	// Concurrent register/unregister on same hub
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			hub.Register(c)
			hub.Unregister(c)
		}(clients[i])
	}
	wg.Wait()

	if count := hub.GetOnlineCount(); count < 0 {
		t.Errorf("invalid online count: %d", count)
	}
}

// TestHubBroadcastWithActiveClients tests broadcast when clients are in rooms
func TestHubBroadcastWithActiveClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Add some fake client entries directly to hub
	hub.mu.Lock()
	for i := 0; i < 5; i++ {
		hub.clients[int64(i)] = &Client{UserID: int64(i), Hub: hub, SendCh: make(chan []byte, 256)}
	}
	hub.mu.Unlock()

	hub.roomMu.Lock()
	hub.rooms[1] = make(map[*Client]bool)
	for i := 0; i < 5; i++ {
		hub.rooms[1][hub.clients[int64(i)]] = true
	}
	hub.roomMu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			hub.BroadcastToRoom(1, 999, []byte("message"), nil)
		}(i)
	}
	wg.Wait()
}
