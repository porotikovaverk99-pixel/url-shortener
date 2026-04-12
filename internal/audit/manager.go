package audit

import (
	"log"
	"sync"
)

type Manager struct {
	observers []Observer
	mu        sync.RWMutex
	logger    Logger
}

type Logger interface {
	Printf(format string, v ...interface{})
}

func NewManager() *Manager {
	return &Manager{
		observers: make([]Observer, 0),
		logger:    log.Default(),
	}
}

func (m *Manager) SetLogger(logger Logger) {
	m.logger = logger
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
		go func(obs Observer, ev Event) {
			if err := obs.Send(ev); err != nil && m.logger != nil {
				m.logger.Printf("audit observer failed to send event: %v", err)
			}
		}(observer, event)
	}
}
