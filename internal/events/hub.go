package events

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	mu       sync.RWMutex
	channels map[string]map[chan Event]struct{}
}

type Event struct {
	ProjectID string          `json:"project_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{channels: make(map[string]map[chan Event]struct{})}
}

func (h *Hub) Subscribe(projectID string) chan Event {
	ch := make(chan Event, 64)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.channels[projectID] == nil {
		h.channels[projectID] = make(map[chan Event]struct{})
	}
	h.channels[projectID][ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(projectID string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if subs, ok := h.channels[projectID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(h.channels, projectID)
		}
	}
	close(ch)
}

func (h *Hub) Publish(projectID string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.channels[projectID] {
		select {
		case ch <- event:
		default:
		}
	}
}
