package buffer

import (
	"sync"
	"time"
)

type Message struct {
	Data      []byte
	Timestamp time.Time
	VIN       string
}

type Buffer struct {
	mu       sync.Mutex
	messages []Message
	capacity int
}

func New(capacity int) *Buffer {
	return &Buffer{
		messages: make([]Message, 0, capacity),
		capacity: capacity,
	}
}

func (b *Buffer) Add(msg Message) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.messages) >= b.capacity {
		return false
	}
	b.messages = append(b.messages, msg)
	return true
}

func (b *Buffer) Drain() []Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.messages) == 0 {
		return nil
	}

	msgs := b.messages
	b.messages = make([]Message, 0, b.capacity)
	return msgs
}

func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.messages)
}


