package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClusterBroadcasterHandleUsersEvent(t *testing.T) {
	hub := NewHub()
	client := &Client{
		UserID: 42,
		SendCh: make(chan []byte, 1),
		done:   make(chan struct{}),
	}
	hub.clients[client.UserID] = client

	broadcaster := NewClusterBroadcaster(nil, hub, "")
	event := clusterEvent{
		Kind:    "users",
		Origin:  "remote-instance",
		UserIDs: []int64{client.UserID},
		Content: []byte("hello"),
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	require.NoError(t, broadcaster.handle(string(payload)))
	select {
	case msg := <-client.SendCh:
		require.Equal(t, []byte("hello"), msg)
	default:
		t.Fatal("expected user fanout message")
	}
}

func TestClusterBroadcasterHandleRoomEvent(t *testing.T) {
	hub := NewHub()
	client := &Client{
		UserID: 7,
		SendCh: make(chan []byte, 1),
		done:   make(chan struct{}),
	}
	hub.rooms[99] = map[*Client]bool{client: true}

	broadcaster := NewClusterBroadcaster(nil, hub, "")
	event := clusterEvent{
		Kind:    "room",
		Origin:  "remote-instance",
		ConvID:  99,
		Content: []byte("room-update"),
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	require.NoError(t, broadcaster.handle(string(payload)))
	select {
	case msg := <-client.SendCh:
		require.Equal(t, []byte("room-update"), msg)
	default:
		t.Fatal("expected room broadcast message")
	}
}

func TestClusterBroadcasterIgnoresOwnEvents(t *testing.T) {
	hub := NewHub()
	client := &Client{
		UserID: 9,
		SendCh: make(chan []byte, 1),
		done:   make(chan struct{}),
	}
	hub.clients[client.UserID] = client

	broadcaster := NewClusterBroadcaster(nil, hub, "")
	event := clusterEvent{
		Kind:    "users",
		Origin:  broadcaster.instanceID,
		UserIDs: []int64{client.UserID},
		Content: []byte("ignored"),
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	require.NoError(t, broadcaster.handle(string(payload)))
	select {
	case <-client.SendCh:
		t.Fatal("did not expect self-origin event to be re-delivered")
	default:
	}
}
