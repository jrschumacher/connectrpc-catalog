package session

import (
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {
	id1, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID failed: %v", err)
	}

	id2, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID failed: %v", err)
	}

	// IDs should be hex encoded, length = SessionIDLength * 2
	expectedLen := SessionIDLength * 2
	if len(id1) != expectedLen {
		t.Errorf("Expected ID length %d, got %d", expectedLen, len(id1))
	}

	// IDs should be unique
	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
}

func TestGetOrCreate(t *testing.T) {
	manager := NewManager(DefaultSessionTTL)
	defer manager.Close()

	// Test creating a new session
	state1, id1, err := manager.GetOrCreate("")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	if state1 == nil {
		t.Fatal("State should not be nil")
	}

	if state1.Registry == nil {
		t.Error("Registry should not be nil")
	}

	if state1.Invoker == nil {
		t.Error("Invoker should not be nil")
	}

	// Test getting existing session
	state2, id2, err := manager.GetOrCreate(id1)
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same session ID, got %s and %s", id1, id2)
	}

	if state1 != state2 {
		t.Error("Expected same state instance")
	}

	// Test creating a new session when ID doesn't exist
	state3, id3, err := manager.GetOrCreate("nonexistent")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	if id3 == id1 {
		t.Error("Should create new session for nonexistent ID")
	}

	if state3 == state1 {
		t.Error("Should create new state for nonexistent ID")
	}
}

func TestGet(t *testing.T) {
	manager := NewManager(DefaultSessionTTL)
	defer manager.Close()

	// Create a session
	state1, id1, err := manager.GetOrCreate("")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	// Get the session
	state2 := manager.Get(id1)
	if state2 == nil {
		t.Fatal("Get should return state")
	}

	if state1 != state2 {
		t.Error("Expected same state instance")
	}

	// Get nonexistent session
	state3 := manager.Get("nonexistent")
	if state3 != nil {
		t.Error("Get should return nil for nonexistent session")
	}
}

func TestDelete(t *testing.T) {
	manager := NewManager(DefaultSessionTTL)
	defer manager.Close()

	// Create a session
	_, id, err := manager.GetOrCreate("")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	// Verify it exists
	if manager.Get(id) == nil {
		t.Fatal("Session should exist")
	}

	// Delete it
	manager.Delete(id)

	// Verify it's gone
	if manager.Get(id) != nil {
		t.Error("Session should be deleted")
	}
}

func TestCleanup(t *testing.T) {
	// Use short TTL for testing
	shortTTL := 100 * time.Millisecond
	manager := NewManager(shortTTL)
	defer manager.Close()

	// Create a session
	state, id, err := manager.GetOrCreate("")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	// Verify it exists
	if manager.Get(id) == nil {
		t.Fatal("Session should exist")
	}

	// Set LastUsed to past
	manager.mu.Lock()
	state.LastUsed = time.Now().Add(-2 * shortTTL)
	manager.mu.Unlock()

	// Run cleanup
	manager.cleanup()

	// Verify it's gone
	if manager.Get(id) != nil {
		t.Error("Expired session should be cleaned up")
	}
}

func TestGetStats(t *testing.T) {
	manager := NewManager(DefaultSessionTTL)
	defer manager.Close()

	// Initially no sessions
	stats := manager.GetStats()
	if stats.ActiveSessions != 0 {
		t.Errorf("Expected 0 active sessions, got %d", stats.ActiveSessions)
	}

	// Create some sessions
	_, _, err := manager.GetOrCreate("")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, _, err = manager.GetOrCreate("")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	stats = manager.GetStats()
	if stats.ActiveSessions != 2 {
		t.Errorf("Expected 2 active sessions, got %d", stats.ActiveSessions)
	}

	if stats.OldestSession == 0 {
		t.Error("OldestSession should be non-zero")
	}

	if stats.NewestSession == 0 {
		t.Error("NewestSession should be non-zero")
	}

	if stats.OldestSession <= stats.NewestSession {
		t.Error("OldestSession should be greater than NewestSession")
	}
}

func TestClose(t *testing.T) {
	manager := NewManager(DefaultSessionTTL)

	// Create some sessions
	_, id1, _ := manager.GetOrCreate("")
	_, id2, _ := manager.GetOrCreate("")

	// Verify they exist
	if manager.Get(id1) == nil || manager.Get(id2) == nil {
		t.Fatal("Sessions should exist")
	}

	// Close the manager
	manager.Close()

	// Verify all sessions are gone
	stats := manager.GetStats()
	if stats.ActiveSessions != 0 {
		t.Errorf("Expected 0 sessions after close, got %d", stats.ActiveSessions)
	}
}

func TestConcurrentAccess(t *testing.T) {
	manager := NewManager(DefaultSessionTTL)
	defer manager.Close()

	// Create a session
	_, id, err := manager.GetOrCreate("")
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				state := manager.Get(id)
				if state == nil {
					t.Error("State should not be nil")
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
