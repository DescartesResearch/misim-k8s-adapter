package inmemorystorage

type InMemBuffer[T any] struct {
	buffer []T
}

func (b *InMemBuffer[T]) Size() int {
	return len(b.buffer)
}

func (b *InMemBuffer[T]) Empty() bool {
	return len(b.buffer) == 0
}

func (b *InMemBuffer[T]) Items() []T {
	// Create a copy of the actual buffer
	cpy := make([]T, len(b.buffer))
	copy(cpy, b.buffer)
	return cpy
}

func (b *InMemBuffer[T]) Clear() []T {
	cpy := b.Items()
	b.buffer = make([]T, 0)
	return cpy
}

func (b *InMemBuffer[T]) Put(newItem T) {
	b.buffer = append(b.buffer, newItem)
}

func (b *InMemBuffer[T]) PutAll(newItems []T) {
	for _, item := range newItems {
		b.Put(item)
	}
}

func NewInMemBuffer[T any]() InMemBuffer[T] {
	return InMemBuffer[T]{
		buffer: make([]T, 0),
	}
}
