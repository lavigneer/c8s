package storage

import (
	"sync"
)

const (
	// MaxBufferSize is the maximum size of the circular buffer (10MB)
	MaxBufferSize = 10 * 1024 * 1024
)

// CircularBuffer is a thread-safe circular buffer for log streaming
type CircularBuffer struct {
	mu          sync.RWMutex
	data        []byte
	size        int
	writePos    int
	wrapped     bool
	subscribers []chan []byte
}

// NewCircularBuffer creates a new circular buffer with the specified size
func NewCircularBuffer(size int) *CircularBuffer {
	if size <= 0 {
		size = MaxBufferSize
	}
	return &CircularBuffer{
		data:        make([]byte, size),
		size:        size,
		subscribers: make([]chan []byte, 0),
	}
}

// Write appends data to the buffer. If the buffer is full, it overwrites oldest data.
func (cb *CircularBuffer) Write(data []byte) (int, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	n := len(data)
	if n == 0 {
		return 0, nil
	}

	// Notify subscribers of new data
	dataCopy := make([]byte, n)
	copy(dataCopy, data)
	for _, ch := range cb.subscribers {
		select {
		case ch <- dataCopy:
		default:
			// Skip if subscriber's channel is full (slow consumer)
		}
	}

	// If data is larger than buffer, only keep the last 'size' bytes
	if n >= cb.size {
		copy(cb.data, data[n-cb.size:])
		cb.writePos = 0
		cb.wrapped = true
		return n, nil
	}

	// Write data to buffer
	for i := 0; i < n; i++ {
		cb.data[cb.writePos] = data[i]
		cb.writePos++
		if cb.writePos >= cb.size {
			cb.writePos = 0
			cb.wrapped = true
		}
	}

	return n, nil
}

// Read returns all data currently in the buffer
func (cb *CircularBuffer) Read() []byte {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if !cb.wrapped {
		// Buffer hasn't wrapped yet, return from start to writePos
		result := make([]byte, cb.writePos)
		copy(result, cb.data[:cb.writePos])
		return result
	}

	// Buffer has wrapped, return writePos to end, then start to writePos
	result := make([]byte, cb.size)
	copy(result, cb.data[cb.writePos:])
	copy(result[cb.size-cb.writePos:], cb.data[:cb.writePos])
	return result
}

// Subscribe creates a channel that receives new log data as it's written
func (cb *CircularBuffer) Subscribe() <-chan []byte {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Create buffered channel to avoid blocking writes
	ch := make(chan []byte, 100)
	cb.subscribers = append(cb.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscription channel and closes it
func (cb *CircularBuffer) Unsubscribe(ch <-chan []byte) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Find and remove the channel
	for i, subscriber := range cb.subscribers {
		if subscriber == ch {
			close(subscriber)
			cb.subscribers = append(cb.subscribers[:i], cb.subscribers[i+1:]...)
			return
		}
	}
}

// Len returns the current number of bytes in the buffer
func (cb *CircularBuffer) Len() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if !cb.wrapped {
		return cb.writePos
	}
	return cb.size
}

// Reset clears the buffer
func (cb *CircularBuffer) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.writePos = 0
	cb.wrapped = false
	cb.data = make([]byte, cb.size)
}
