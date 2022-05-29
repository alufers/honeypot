package main

import "sync"

type EventBroadcaster[T any] struct {
	listeners      []chan T
	listenersMutex sync.RWMutex
}

func NewEventBroadcaster[T any]() *EventBroadcaster[T] {
	return &EventBroadcaster[T]{
		listeners: []chan T{},
	}
}

func (b *EventBroadcaster[T]) AddListener(listener chan T) (cancelFunc func()) {
	b.listenersMutex.Lock()
	defer b.listenersMutex.Unlock()
	b.listeners = append(b.listeners, listener)
	return func() {
		b.listenersMutex.Lock()
		defer b.listenersMutex.Unlock()
		for i, l := range b.listeners {
			if l == listener {
				b.listeners = append(b.listeners[:i], b.listeners[i+1:]...)
				return
			}
		}
	}
}

func (b *EventBroadcaster[T]) Broadcast(event T) {
	b.listenersMutex.RLock()
	defer b.listenersMutex.RUnlock()
	for _, listener := range b.listeners {
		select {
		case listener <- event:
		default:
		}
	}
}
