package inventory

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Update struct {
	ProductID uuid.UUID `json:"productId"`
	NewStock  int       `json:"newStock"`
	Delta     int       `json:"delta"`
	Timestamp time.Time `json:"timestamp"`
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]chan Update
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]chan Update),
	}
}

func (h *Hub) Subscribe(id string) <-chan Update {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan Update, 64)
	h.subscribers[id] = ch
	return ch
}

func (h *Hub) Unsubscribe(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if ch, ok := h.subscribers[id]; ok {
		close(ch)
		delete(h.subscribers, id)
	}
}

func (h *Hub) Publish(update Update) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.subscribers {
		select {
		case ch <- update:
		default:
		}
	}
}

func (h *Hub) SubscriberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers)
}
