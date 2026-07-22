package event

import (
	"sync"
	"sync/atomic"
)

// Bus is a simple in-process pub/sub for events.
// Publishes are non-blocking: slow subscribers drop events rather than stall producers.
type Bus struct {
	mu   sync.RWMutex
	subs map[uint64]chan Event
	next atomic.Uint64
}

// NewBus constructs an empty event bus.
func NewBus() *Bus {
	return &Bus{
		subs: make(map[uint64]chan Event),
	}
}

// Subscribe registers a buffered channel that receives published events.
// The returned cancel function removes the subscription and closes the channel.
func (b *Bus) Subscribe() (<-chan Event, func()) {
	id := b.next.Add(1)
	ch := make(chan Event, 64)

	b.mu.Lock()
	b.subs[id] = ch
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		if existing, ok := b.subs[id]; ok {
			delete(b.subs, id)
			close(existing)
		}
		b.mu.Unlock()
	}
	return ch, cancel
}

// Publish fans out e to all subscribers without blocking.
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- e:
		default:
			// Drop if subscriber is slow — never block producers.
		}
	}
}
