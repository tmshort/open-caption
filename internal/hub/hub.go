// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Package hub fans pre-serialized server-sent-event frames out to audience
// clients without ever blocking the publisher (docs/02-architecture.md, decision 8).
package hub

import (
	"fmt"
	"sort"
	"sync"
	"time"
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
	C    chan Frame
	info ClientInfo
}

// ClientInfo says who a subscriber is, for the list of connected devices.
type ClientInfo struct {
	IP        string    `json:"ip"`
	UserAgent string    `json:"-"`
	Projector bool      `json:"projector,omitempty"` // opened as a projector / shared screen (?display=projector)
	Since     time.Time `json:"since"`
}

// Hub tracks subscribers.
type Hub struct {
	mu      sync.Mutex
	clients map[*Client]struct{}
	max     int
	dropped int64

	// OnChange, if set before use, is called (outside the hub lock) after the
	// client count changes. It must not block.
	OnChange func()
}

// New creates a hub accepting at most max clients (<=0: unlimited).
func New(max int) *Hub { return &Hub{clients: map[*Client]struct{}{}, max: max} }

// Subscribe registers a client, or returns nil if the hub is full.
func (h *Hub) Subscribe() *Client { return h.SubscribeInfo(ClientInfo{}) }

// SubscribeInfo is Subscribe for a client that can say who it is.
func (h *Hub) SubscribeInfo(info ClientInfo) *Client {
	if info.Since.IsZero() {
		info.Since = time.Now()
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.max > 0 && len(h.clients) >= h.max {
		return nil
	}
	c := &Client{C: make(chan Frame, bufSize), info: info}
	h.clients[c] = struct{}{}
	h.changed()
	return c
}

// changed must be called with h.mu held; the callback runs asynchronously so
// it can never deadlock against the hub.
func (h *Hub) changed() {
	if h.OnChange != nil {
		go h.OnChange()
	}
}

// Unsubscribe removes a client. It is safe to call more than once.
func (h *Hub) Unsubscribe(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.C)
		h.changed()
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
			h.changed()
		}
	}
}

// Clients lists who is connected, oldest first.
func (h *Hub) Clients() []ClientInfo {
	h.mu.Lock()
	out := make([]ClientInfo, 0, len(h.clients))
	for c := range h.clients {
		out = append(out, c.info)
	}
	h.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Since.Before(out[j].Since) })
	return out
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
