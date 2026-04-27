package sse

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type client struct {
	id string
	ch chan []byte
}

type Broker struct {
	clients    map[string]chan []byte
	register   chan client
	unregister chan string
	broadcast  chan []byte
	mu         sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients:    make(map[string]chan []byte),
		register:   make(chan client),
		unregister: make(chan string),
		broadcast:  make(chan []byte, 16),
	}
}

func (b *Broker) Subscribe(userID string) (<-chan []byte, func()) {
	clientID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
	ch := make(chan []byte, 8)
	b.register <- client{id: clientID, ch: ch}

	unsubscribe := func() {
		b.unregister <- clientID
	}

	return ch, unsubscribe
}

func (b *Broker) Publish(event string, data any) {
	payload, err := json.Marshal(map[string]any{
		"event": event,
		"data":  data,
	})
	if err != nil {
		log.Error().Err(err).Str("event", event).Msg("failed to marshal sse payload")
		return
	}

	b.broadcast <- payload
}

func (b *Broker) Run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client.id] = client.ch
			b.mu.Unlock()
		case clientID := <-b.unregister:
			b.mu.Lock()
			if ch, ok := b.clients[clientID]; ok {
				delete(b.clients, clientID)
				close(ch)
			}
			b.mu.Unlock()
		case payload := <-b.broadcast:
			b.mu.RLock()
			for clientID, ch := range b.clients {
				select {
				case ch <- payload:
				default:
					log.Warn().Str("client_id", clientID).Msg("dropping slow sse client")
				}
			}
			b.mu.RUnlock()
		}
	}
}
