package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/opentdf/connectrpc-catalog/internal/invoker"
	"github.com/opentdf/connectrpc-catalog/internal/registry"
)

const (
	// DefaultSessionTTL is the default time-to-live for sessions
	DefaultSessionTTL = 1 * time.Hour
	// CleanupInterval is how often to check for expired sessions
	CleanupInterval = 5 * time.Minute
	// SessionIDLength is the length of session IDs in bytes (will be hex encoded)
	SessionIDLength = 16
)

// State holds the per-session state
type State struct {
	Registry  *registry.Registry
	Invoker   *invoker.Invoker
	CreatedAt time.Time
	LastUsed  time.Time
}

// Manager handles session lifecycle
type Manager struct {
	sessions map[string]*State
	mu       sync.RWMutex
	ttl      time.Duration
	stopCh   chan struct{}
}

// NewManager creates a new session manager
func NewManager(ttl time.Duration) *Manager {
	if ttl <= 0 {
		ttl = DefaultSessionTTL
	}

	m := &Manager{
		sessions: make(map[string]*State),
		ttl:      ttl,
		stopCh:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go m.cleanupLoop()

	return m
}

// GenerateID creates a new random session ID
func GenerateID() (string, error) {
	bytes := make([]byte, SessionIDLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetOrCreate returns an existing session or creates a new one
func (m *Manager) GetOrCreate(sessionID string) (*State, string, error) {
	// Try to get existing session
	if sessionID != "" {
		m.mu.RLock()
		state, exists := m.sessions[sessionID]
		m.mu.RUnlock()

		if exists {
			m.mu.Lock()
			state.LastUsed = time.Now()
			m.mu.Unlock()
			return state, sessionID, nil
		}
	}

	// Create new session
	newID, err := GenerateID()
	if err != nil {
		return nil, "", err
	}

	state := &State{
		Registry:  registry.New(),
		Invoker:   invoker.New(),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	m.mu.Lock()
	m.sessions[newID] = state
	m.mu.Unlock()

	return state, newID, nil
}

// Get returns a session by ID, or nil if not found
func (m *Manager) Get(sessionID string) *State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.sessions[sessionID]
	if !exists {
		return nil
	}

	// Update last used time
	state.LastUsed = time.Now()
	return state
}

// Delete removes a session
func (m *Manager) Delete(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if state, exists := m.sessions[sessionID]; exists {
		if state.Invoker != nil {
			state.Invoker.Close()
		}
		delete(m.sessions, sessionID)
	}
}

// cleanupLoop periodically removes expired sessions
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// cleanup removes expired sessions
func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, state := range m.sessions {
		if now.Sub(state.LastUsed) > m.ttl {
			if state.Invoker != nil {
				state.Invoker.Close()
			}
			delete(m.sessions, id)
		}
	}
}

// Close stops the cleanup loop and cleans up all sessions
func (m *Manager) Close() {
	close(m.stopCh)

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, state := range m.sessions {
		if state.Invoker != nil {
			state.Invoker.Close()
		}
		delete(m.sessions, id)
	}
}

// Stats returns session statistics
type Stats struct {
	ActiveSessions int
	OldestSession  time.Duration
	NewestSession  time.Duration
}

// GetStats returns current session statistics
func (m *Manager) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		ActiveSessions: len(m.sessions),
	}

	now := time.Now()
	for _, state := range m.sessions {
		age := now.Sub(state.CreatedAt)
		if stats.OldestSession == 0 || age > stats.OldestSession {
			stats.OldestSession = age
		}
		if stats.NewestSession == 0 || age < stats.NewestSession {
			stats.NewestSession = age
		}
	}

	return stats
}
