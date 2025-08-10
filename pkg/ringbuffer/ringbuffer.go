package ringbuffer

import "sync"

type RingBuffer[T any] struct {
	mutex    sync.RWMutex
	capacity int
	head     int
	buffer   []T
}

func NewBuffer[T any](capacity int) *RingBuffer[T] {
	if capacity < 1 {
		capacity = 1
	}

	return &RingBuffer[T]{
		buffer:   make([]T, 0, capacity),
		capacity: capacity,
	}
}

func (b *RingBuffer[T]) Add(message T) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.buffer) < b.capacity {
		b.buffer = append(b.buffer, message)
		return
	}

	b.buffer[b.head] = message
	b.head = (b.head + 1) % b.capacity
}

func (b *RingBuffer[T]) Get() []T {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	result := make([]T, len(b.buffer))

	if len(b.buffer) < b.capacity {
		copy(result, b.buffer)
	} else {
		n := copy(result, b.buffer[b.head:])
		copy(result[n:], b.buffer[:b.head])
	}

	return result
}

func (b *RingBuffer[T]) Len() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return len(b.buffer)
}
