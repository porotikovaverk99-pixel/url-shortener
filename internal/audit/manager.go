package audit

import "sync"

type Manager struct {
	observers []Observer
	mu        sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		observers: make([]Observer, 0),
	}
}

func (m *Manager) AddObserver(observer Observer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observers = append(m.observers, observer)
}

func (m *Manager) Notify(event Event) {
	m.mu.RLock()
	observers := make([]Observer, len(m.observers))
	copy(observers, m.observers)
	m.mu.RUnlock()

	for _, observer := range observers {
		go observer.Send(event)
	}
}
