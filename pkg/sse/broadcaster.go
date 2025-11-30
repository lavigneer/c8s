package sse

import (
	"fmt"
	"sync"
)

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	ID    string
	Event string
	Data  string
}

// String formats the event as SSE format
func (e SSEEvent) String() string {
	result := ""
	if e.ID != "" {
		result += fmt.Sprintf("id: %s\n", e.ID)
	}
	if e.Event != "" {
		result += fmt.Sprintf("event: %s\n", e.Event)
	}
	if e.Data != "" {
		result += fmt.Sprintf("data: %s\n", e.Data)
	}
	result += "\n"
	return result
}

// Broadcaster implements pub/sub for broadcasting SSE events to multiple clients
type Broadcaster struct {
	clients map[chan SSEEvent]bool
	mutex   sync.RWMutex
	done    chan struct{}
}

// NewBroadcaster creates a new SSE broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan SSEEvent]bool),
		done:    make(chan struct{}),
	}
}

// Subscribe adds a new subscriber channel
func (b *Broadcaster) Subscribe() chan SSEEvent {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	ch := make(chan SSEEvent, 10) // Buffered channel to prevent blocking
	b.clients[ch] = true
	return ch
}

// Unsubscribe removes a subscriber channel
func (b *Broadcaster) Unsubscribe(ch chan SSEEvent) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
}

// Broadcast sends an event to all subscribers
func (b *Broadcaster) Broadcast(event SSEEvent) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	for ch := range b.clients {
		select {
		case ch <- event:
		case <-b.done:
			return
		default:
			// Channel buffer full, skip this client
		}
	}
}

// BroadcastAsync broadcasts event asynchronously
func (b *Broadcaster) BroadcastAsync(event SSEEvent) {
	go b.Broadcast(event)
}

// ClientCount returns number of connected clients
func (b *Broadcaster) ClientCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return len(b.clients)
}

// Close closes the broadcaster and all client channels
func (b *Broadcaster) Close() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	close(b.done)
	for ch := range b.clients {
		close(ch)
	}
	b.clients = make(map[chan SSEEvent]bool)
}

// EventBuilder is a helper for building SSE events
type EventBuilder struct {
	event SSEEvent
}

// NewEventBuilder creates a new event builder
func NewEventBuilder() *EventBuilder {
	return &EventBuilder{}
}

// WithID sets the event ID
func (eb *EventBuilder) WithID(id string) *EventBuilder {
	eb.event.ID = id
	return eb
}

// WithEvent sets the event type
func (eb *EventBuilder) WithEvent(eventType string) *EventBuilder {
	eb.event.Event = eventType
	return eb
}

// WithData sets the event data
func (eb *EventBuilder) WithData(data string) *EventBuilder {
	eb.event.Data = data
	return eb
}

// Build returns the constructed event
func (eb *EventBuilder) Build() SSEEvent {
	return eb.event
}

// Helper functions for common event types

// NewStatusUpdateEvent creates a status update event
func NewStatusUpdateEvent(id, status, message string) SSEEvent {
	return SSEEvent{
		ID:    id,
		Event: "status",
		Data:  fmt.Sprintf(`{"status":%q,"message":%q}`, status, message),
	}
}

// NewLogEvent creates a log event
func NewLogEvent(id, message string) SSEEvent {
	return SSEEvent{
		ID:    id,
		Event: "log",
		Data:  fmt.Sprintf(`{"message":%q}`, message),
	}
}

// NewProgressEvent creates a progress event
func NewProgressEvent(id string, progress int) SSEEvent {
	return SSEEvent{
		ID:    id,
		Event: "progress",
		Data:  fmt.Sprintf(`{"progress":%d}`, progress),
	}
}

// NewErrorEvent creates an error event
func NewErrorEvent(id, errorMsg string) SSEEvent {
	return SSEEvent{
		ID:    id,
		Event: "error",
		Data:  fmt.Sprintf(`{"error":%q}`, errorMsg),
	}
}
