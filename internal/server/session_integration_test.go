package server

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
)

func TestSessionCreation(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create first session by making a request without session ID
	req1 := connect.NewRequest(&catalogv1.ListServicesRequest{})
	resp1, err := server.ListServices(ctx, req1)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	sessionID1 := resp1.Header().Get("X-Session-ID")
	if sessionID1 == "" {
		t.Fatal("Expected X-Session-ID header in response")
	}

	// Initially no services
	if len(resp1.Msg.Services) != 0 {
		t.Error("Expected 0 services initially")
	}

	// Create second session
	req2 := connect.NewRequest(&catalogv1.ListServicesRequest{})
	resp2, err := server.ListServices(ctx, req2)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	sessionID2 := resp2.Header().Get("X-Session-ID")
	if sessionID2 == "" {
		t.Fatal("Expected X-Session-ID header in response")
	}

	// Should be different session
	if sessionID1 == sessionID2 {
		t.Error("Expected different session IDs")
	}

	// Both sessions should have no services
	if len(resp2.Msg.Services) != 0 {
		t.Error("Expected 0 services in second session")
	}
}

func TestSessionPersistence(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create session
	req1 := connect.NewRequest(&catalogv1.ListServicesRequest{})
	resp1, err := server.ListServices(ctx, req1)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	sessionID := resp1.Header().Get("X-Session-ID")
	if sessionID == "" {
		t.Fatal("Expected X-Session-ID header")
	}

	// Make another request with the same session ID
	req2 := connect.NewRequest(&catalogv1.ListServicesRequest{})
	req2.Header().Set("X-Session-ID", sessionID)
	resp2, err := server.ListServices(ctx, req2)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	returnedSessionID := resp2.Header().Get("X-Session-ID")
	if returnedSessionID != sessionID {
		t.Errorf("Expected same session ID, got %s", returnedSessionID)
	}
}

func TestMultipleRequestsSameSession(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create session
	req1 := connect.NewRequest(&catalogv1.ListServicesRequest{})
	resp1, err := server.ListServices(ctx, req1)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	sessionID := resp1.Header().Get("X-Session-ID")

	// Make multiple requests with the same session
	for i := 0; i < 5; i++ {
		req := connect.NewRequest(&catalogv1.ListServicesRequest{})
		req.Header().Set("X-Session-ID", sessionID)
		resp, err := server.ListServices(ctx, req)
		if err != nil {
			t.Fatalf("ListServices failed: %v", err)
		}

		returnedID := resp.Header().Get("X-Session-ID")
		if returnedID != sessionID {
			t.Errorf("Request %d: expected session ID %s, got %s", i, sessionID, returnedID)
		}
	}

	// Verify session still exists
	stats := server.GetStats()
	if stats.SessionStats.ActiveSessions != 1 {
		t.Errorf("Expected 1 active session, got %d", stats.SessionStats.ActiveSessions)
	}
}

func TestInvalidSessionID(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Use invalid/nonexistent session ID
	req := connect.NewRequest(&catalogv1.ListServicesRequest{})
	req.Header().Set("X-Session-ID", "invalid-session-id")
	resp, err := server.ListServices(ctx, req)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	// Should create new session
	newSessionID := resp.Header().Get("X-Session-ID")
	if newSessionID == "" {
		t.Fatal("Expected X-Session-ID header")
	}

	if newSessionID == "invalid-session-id" {
		t.Error("Should create new session for invalid ID")
	}
}

func TestSessionStatsTracking(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Initially no sessions
	stats := server.GetStats()
	if stats.SessionStats.ActiveSessions != 0 {
		t.Errorf("Expected 0 active sessions, got %d", stats.SessionStats.ActiveSessions)
	}

	// Create multiple sessions
	var sessionIDs []string
	for i := 0; i < 3; i++ {
		req := connect.NewRequest(&catalogv1.ListServicesRequest{})
		resp, err := server.ListServices(ctx, req)
		if err != nil {
			t.Fatalf("ListServices failed: %v", err)
		}
		sessionIDs = append(sessionIDs, resp.Header().Get("X-Session-ID"))
	}

	// Check stats
	stats = server.GetStats()
	if stats.SessionStats.ActiveSessions != 3 {
		t.Errorf("Expected 3 active sessions, got %d", stats.SessionStats.ActiveSessions)
	}

	if stats.SessionStats.OldestSession == 0 {
		t.Error("OldestSession should be non-zero")
	}

	if stats.SessionStats.NewestSession == 0 {
		t.Error("NewestSession should be non-zero")
	}
}

func TestGetServiceSchemaWithSession(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create session
	req1 := connect.NewRequest(&catalogv1.GetServiceSchemaRequest{
		ServiceName: "test.Service",
	})
	resp1, err := server.GetServiceSchema(ctx, req1)
	if err != nil {
		t.Fatalf("GetServiceSchema failed: %v", err)
	}

	sessionID := resp1.Header().Get("X-Session-ID")
	if sessionID == "" {
		t.Fatal("Expected X-Session-ID header")
	}

	// Service not found (expected since nothing loaded)
	if resp1.Msg.Error == "" {
		t.Error("Expected error for nonexistent service")
	}

	// Make another request with same session
	req2 := connect.NewRequest(&catalogv1.GetServiceSchemaRequest{
		ServiceName: "test.Service",
	})
	req2.Header().Set("X-Session-ID", sessionID)
	resp2, err := server.GetServiceSchema(ctx, req2)
	if err != nil {
		t.Fatalf("GetServiceSchema failed: %v", err)
	}

	returnedID := resp2.Header().Get("X-Session-ID")
	if returnedID != sessionID {
		t.Errorf("Expected same session ID, got %s", returnedID)
	}
}
