package main

import "sync"

type liveReloadHub struct {
	mu          sync.Mutex
	subscribers map[chan struct{}]struct{}
}

func newLiveReloadHub() *liveReloadHub {
	return &liveReloadHub{
		subscribers: make(map[chan struct{}]struct{}),
	}
}

func (h *liveReloadHub) Subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *liveReloadHub) Unsubscribe(ch chan struct{}) {
	h.mu.Lock()
	delete(h.subscribers, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *liveReloadHub) Publish() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subscribers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
