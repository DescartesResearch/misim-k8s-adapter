package broadcast

import (
	"context"
	"k8s.io/klog/v2"
)

// https://betterprogramming.pub/how-to-broadcast-messages-in-go-using-channels-b68f42bdf32e

type BroadcastServer[T any] struct {
	source         <-chan T
	listeners      []chan T
	addListener    chan chan T
	removeListener chan (<-chan T)
	name           string
}

func (s *BroadcastServer[T]) Subscribe() <-chan T {
	klog.V(7).Info("Subscribe to ", s.name)
	newListener := make(chan T, 500)
	s.addListener <- newListener
	return newListener
}

func (s *BroadcastServer[T]) CancelSubscription(channel <-chan T) {
	klog.V(7).Info("Remove from ", s.name)
	s.removeListener <- channel
}

func NewBroadcastServer[T any](ctx context.Context, name string, source <-chan T) *BroadcastServer[T] {
	service := &BroadcastServer[T]{
		source:         source,
		listeners:      make([]chan T, 0),
		addListener:    make(chan chan T),
		removeListener: make(chan (<-chan T)),
		name:           name,
	}
	go service.serve(ctx)
	return service
}

func (s *BroadcastServer[T]) serve(ctx context.Context) {
	defer func() {
		for _, listener := range s.listeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addListener:
			s.listeners = append(s.listeners, newListener)
		case listenerToRemove := <-s.removeListener:
			for i, ch := range s.listeners {
				if ch == listenerToRemove {
					s.listeners[i] = s.listeners[len(s.listeners)-1]
					s.listeners = s.listeners[:len(s.listeners)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-s.source:
			if !ok {
				return
			}
			for _, listener := range s.listeners {
				if listener != nil {
					select {
					case listener <- val:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}
