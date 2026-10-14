// Package hub fans pre-serialized server-sent-event frames out to audience
// clients without ever blocking the publisher (docs/02-architecture.md, decision 8).
package hub

import (
	"fmt"
	"sync"
)

// Frame is one SSE message. Seq is 0 for frames that carry no sequence number
// (hints such as preload); those are always delivered.
type Frame struct {
	Seq  int64
	Data []byte
}

// FormatEvent renders an SSE event. id < 0 omits the id line.
func FormatEvent(id int64, typ string, data []byte) []byte {
	if id >= 0 {
		return []byte(fmt.Sprintf("id: %d\nevent: %s\ndata: %s\n\n", id, typ, data))
	}
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", typ, data))
}

const bufSize = 64

// Client is one subscriber. C is closed if the client was dropped for being
// too slow or the hub was closed; the handler should then end the response
// so the browser reconnects and receives a fresh snapshot.
type Client struct {
	C chan Frame
}

// Hub tracks subscribers.
type Hub struct {
	mu      sync.Mutex
	clients map[*Client]struct{}
	max     int
	dropped int64
}

// New creates a hub accepting at most max clients (<=0: unlimited).
func New(max int) *Hub { return &Hub{clients: map[*Client]struct{}{}, max: max} }

// Subscribe registers a client, or returns nil if the hub is full.
func (h *Hub) Subscribe() *Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.max > 0 && len(h.clients) >= h.max {
		return nil
	}
	c := &Client{C: make(chan Frame, bufSize)}
	h.clients[c] = struct{}{}
	return c
}

// Unsubscribe removes a client. It is safe to call more than once.
func (h *Hub) Unsubscribe(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.C)
	}
}

// Broadcast delivers f to every client. Clients whose buffer is full are dropped.
func (h *Hub) Broadcast(f Frame) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.C <- f:
		default:
			delete(h.clients, c)
			close(c.C)
			h.dropped++
		}
	}
}

// Count returns the number of connected clients.
func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// Dropped returns how many clients were dropped for being too slow.
func (h *Hub) Dropped() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.dropped
}
