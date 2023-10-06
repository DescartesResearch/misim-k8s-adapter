package inmemorystorage

type InMemChannelWrapper[T any] struct {
	channel chan T
}

func (w *InMemChannelWrapper[T]) InitChannel() chan T {
	w.channel = make(chan T)
	return w.channel
}

func (w *InMemChannelWrapper[T]) Get() chan T {
	return w.channel
}

func NewInMemChannelWrapper[T any]() InMemChannelWrapper[T] {
	return InMemChannelWrapper[T]{
		channel: make(chan T),
	}
}
