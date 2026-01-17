package server

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
)

// TestLoadProtos tests loading proto files from a local path
// Note: This test is skipped by default as it requires buf CLI and proto files
func TestLoadProtos(t *testing.T) {
	t.Skip("Skipping test that requires buf CLI - use integration tests with real proto files")
}

// TestLoadProtos_WithTestData tests loading protos using programmatically created test data
func TestLoadProtos_WithTestData(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create test file descriptor set
	fds := createTestFileDescriptorSet()

	// Create a session and register it directly (bypassing LoadProtos RPC)
	state, sessionID, err := server.sessionManager.GetOrCreate("")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if err := state.Registry.Register(fds); err != nil {
		t.Fatalf("Failed to register test descriptors: %v", err)
	}

	// Verify registration worked by listing services with session
	req := connect.NewRequest(&catalogv1.ListServicesRequest{})
	req.Header().Set("X-Session-ID", sessionID)
	resp, err := server.ListServices(ctx, req)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	services := resp.Msg.Services
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	if len(services) > 0 && services[0].Name != "test.v1.TestService" {
		t.Errorf("Expected service name 'test.v1.TestService', got '%s'", services[0].Name)
	}
}

// TestLoadProtos_InvalidPath tests error handling for invalid paths
func TestLoadProtos_InvalidPath(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	req := connect.NewRequest(&catalogv1.LoadProtosRequest{
		Source: &catalogv1.LoadProtosRequest_ProtoPath{
			ProtoPath: "/nonexistent/path/to/protos",
		},
	})

	resp, err := server.LoadProtos(ctx, req)
	if err != nil {
		t.Fatalf("LoadProtos returned error: %v", err)
	}

	if resp.Msg.Success != false {
		t.Errorf("Expected success=false for invalid path, got success=%v", resp.Msg.Success)
	}

	if resp.Msg.Error == "" {
		t.Error("Expected error message for invalid path, got empty string")
	}
}

// TestListServices tests listing services after loading protos
func TestListServices(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create a session and register test descriptors directly
	state, sessionID, err := server.sessionManager.GetOrCreate("")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	fds := createTestFileDescriptorSet()
	if err := state.Registry.Register(fds); err != nil {
		t.Fatalf("Failed to register test descriptors: %v", err)
	}

	// Now list services with session
	listReq := connect.NewRequest(&catalogv1.ListServicesRequest{})
	listReq.Header().Set("X-Session-ID", sessionID)

	listResp, err := server.ListServices(ctx, listReq)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	if len(listResp.Msg.Services) == 0 {
		t.Fatal("Expected at least one service, got zero")
	}

	// Verify the TestService is present
	foundTestService := false
	for _, svc := range listResp.Msg.Services {
		if svc.Name == "test.v1.TestService" {
			foundTestService = true

			// Verify it has methods
			if len(svc.Methods) == 0 {
				t.Error("TestService should have methods")
			}

			// Check for TestMethod
			foundTestMethod := false
			for _, method := range svc.Methods {
				if method.Name == "TestMethod" {
					foundTestMethod = true
					// Verify method properties
					if method.InputType != "test.v1.TestRequest" {
						t.Errorf("Expected input type 'test.v1.TestRequest', got '%s'", method.InputType)
					}
					if method.OutputType != "test.v1.TestResponse" {
						t.Errorf("Expected output type 'test.v1.TestResponse', got '%s'", method.OutputType)
					}
				}
			}

			if !foundTestMethod {
				t.Error("Expected method TestMethod not found in TestService")
			}

			break
		}
	}

	if !foundTestService {
		t.Error("Expected to find test.v1.TestService in service list")
	}
}

// TestListServices_Empty tests listing services when none are loaded
func TestListServices_Empty(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	req := connect.NewRequest(&catalogv1.ListServicesRequest{})

	resp, err := server.ListServices(ctx, req)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	if len(resp.Msg.Services) != 0 {
		t.Errorf("Expected zero services for empty registry, got %d", len(resp.Msg.Services))
	}
}

// TestGetServiceSchema tests retrieving service schema
func TestGetServiceSchema(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create a session and register test descriptors directly
	state, sessionID, err := server.sessionManager.GetOrCreate("")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	fds := createTestFileDescriptorSet()
	if err := state.Registry.Register(fds); err != nil {
		t.Fatalf("Failed to register test descriptors: %v", err)
	}

	// Get schema for TestService with session
	schemaReq := connect.NewRequest(&catalogv1.GetServiceSchemaRequest{
		ServiceName: "test.v1.TestService",
	})
	schemaReq.Header().Set("X-Session-ID", sessionID)

	schemaResp, err := server.GetServiceSchema(ctx, schemaReq)
	if err != nil {
		t.Fatalf("GetServiceSchema failed: %v", err)
	}

	if schemaResp.Msg.Error != "" {
		t.Errorf("Expected no error, got: %s", schemaResp.Msg.Error)
	}

	if schemaResp.Msg.Service == nil {
		t.Fatal("Expected service info, got nil")
	}

	if schemaResp.Msg.Service.Name != "test.v1.TestService" {
		t.Errorf("Expected service name 'test.v1.TestService', got '%s'", schemaResp.Msg.Service.Name)
	}

	// Verify message schemas are returned
	if len(schemaResp.Msg.MessageSchemas) == 0 {
		t.Error("Expected message schemas, got zero")
	}

	// Check for expected message types
	expectedMessages := []string{
		"test.v1.TestRequest",
		"test.v1.TestResponse",
	}

	for _, msgName := range expectedMessages {
		if _, exists := schemaResp.Msg.MessageSchemas[msgName]; !exists {
			t.Errorf("Expected message schema for %s not found", msgName)
		}
	}
}

// TestGetServiceSchema_NotFound tests error handling for unknown service
func TestGetServiceSchema_NotFound(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create a session and register test descriptors first
	state, sessionID, err := server.sessionManager.GetOrCreate("")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	fds := createTestFileDescriptorSet()
	if err := state.Registry.Register(fds); err != nil {
		t.Fatalf("Failed to register test descriptors: %v", err)
	}

	// Try to get schema for non-existent service with session
	schemaReq := connect.NewRequest(&catalogv1.GetServiceSchemaRequest{
		ServiceName: "nonexistent.Service",
	})
	schemaReq.Header().Set("X-Session-ID", sessionID)

	schemaResp, err := server.GetServiceSchema(ctx, schemaReq)
	if err != nil {
		t.Fatalf("GetServiceSchema failed: %v", err)
	}

	if schemaResp.Msg.Error == "" {
		t.Error("Expected error for non-existent service, got empty string")
	}
}

// TestGetServiceSchema_EmptyName tests validation for empty service name
func TestGetServiceSchema_EmptyName(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	schemaReq := connect.NewRequest(&catalogv1.GetServiceSchemaRequest{
		ServiceName: "",
	})

	_, err := server.GetServiceSchema(ctx, schemaReq)
	if err == nil {
		t.Error("Expected error for empty service name, got nil")
	}

	// Verify it's an InvalidArgument error
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("Expected InvalidArgument error code, got %v", connect.CodeOf(err))
	}
}

// TestInvokeGRPC tests the InvokeGRPC handler validation
// Note: This test only validates request validation, not actual invocation
// since we don't have a running gRPC server to invoke in unit tests
func TestInvokeGRPC(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create a session and register test descriptors directly
	state, sessionID, err := server.sessionManager.GetOrCreate("")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	fds := createTestFileDescriptorSet()
	if err := state.Registry.Register(fds); err != nil {
		t.Fatalf("Failed to register test descriptors: %v", err)
	}

	// Test invocation with valid request structure and session
	// Note: This will fail to connect since there's no actual server,
	// but we can verify the request validation works
	invokeReq := connect.NewRequest(&catalogv1.InvokeGRPCRequest{
		Endpoint:    "localhost:9999",
		Service:     "test.v1.TestService",
		Method:      "TestMethod",
		RequestJson: `{"name": "test"}`,
		UseTls:      false,
	})
	invokeReq.Header().Set("X-Session-ID", sessionID)

	invokeResp, err := server.InvokeGRPC(ctx, invokeReq)
	if err != nil {
		t.Fatalf("InvokeGRPC failed: %v", err)
	}

	// We expect success=false because there's no server listening
	// but the validation should pass
	if invokeResp.Msg.Success {
		t.Error("Expected success=false (no server running), got success=true")
	}

	// Error should mention connection issues, not validation issues
	if invokeResp.Msg.Error == "" {
		t.Error("Expected connection error message, got empty string")
	}
}

// TestInvokeGRPC_MissingEndpoint tests validation for missing endpoint
func TestInvokeGRPC_MissingEndpoint(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	invokeReq := connect.NewRequest(&catalogv1.InvokeGRPCRequest{
		Service:     "catalog.v1.CatalogService",
		Method:      "ListServices",
		RequestJson: "{}",
	})

	_, err := server.InvokeGRPC(ctx, invokeReq)
	if err == nil {
		t.Error("Expected error for missing endpoint, got nil")
	}

	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("Expected InvalidArgument error code, got %v", connect.CodeOf(err))
	}
}

// TestInvokeGRPC_MissingService tests validation for missing service
func TestInvokeGRPC_MissingService(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	invokeReq := connect.NewRequest(&catalogv1.InvokeGRPCRequest{
		Endpoint:    "localhost:9999",
		Method:      "ListServices",
		RequestJson: "{}",
	})

	_, err := server.InvokeGRPC(ctx, invokeReq)
	if err == nil {
		t.Error("Expected error for missing service, got nil")
	}

	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("Expected InvalidArgument error code, got %v", connect.CodeOf(err))
	}
}

// TestInvokeGRPC_MissingMethod tests validation for missing method
func TestInvokeGRPC_MissingMethod(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	invokeReq := connect.NewRequest(&catalogv1.InvokeGRPCRequest{
		Endpoint:    "localhost:9999",
		Service:     "catalog.v1.CatalogService",
		RequestJson: "{}",
	})

	_, err := server.InvokeGRPC(ctx, invokeReq)
	if err == nil {
		t.Error("Expected error for missing method, got nil")
	}

	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("Expected InvalidArgument error code, got %v", connect.CodeOf(err))
	}
}

// TestServerValidation tests the ValidateSetup method
func TestServerValidation(t *testing.T) {
	server := New()
	defer server.Close()

	if err := server.ValidateSetup(); err != nil {
		t.Errorf("ValidateSetup failed: %v", err)
	}
}

// TestServerStats tests the GetStats method
func TestServerStats(t *testing.T) {
	server := New()
	defer server.Close()

	stats := server.GetStats()

	// Verify stats structure is populated
	if stats.SessionStats.ActiveSessions < 0 {
		t.Error("Expected non-negative active sessions")
	}
}

// TestSessionIsolation tests that sessions are isolated from each other
func TestSessionIsolation(t *testing.T) {
	server := New()
	defer server.Close()

	ctx := context.Background()

	// Create first session and register test descriptors
	state1, sessionID1, err := server.sessionManager.GetOrCreate("")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	fds := createTestFileDescriptorSet()
	if err := state1.Registry.Register(fds); err != nil {
		t.Fatalf("Failed to register test descriptors: %v", err)
	}

	// Verify services are loaded in session 1
	listReq1 := connect.NewRequest(&catalogv1.ListServicesRequest{})
	listReq1.Header().Set("X-Session-ID", sessionID1)
	listResp1, err := server.ListServices(ctx, listReq1)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}
	if len(listResp1.Msg.Services) == 0 {
		t.Fatal("Expected services to be loaded in session 1")
	}

	// Create second session (should have no services)
	listReq2 := connect.NewRequest(&catalogv1.ListServicesRequest{})
	listResp2, err := server.ListServices(ctx, listReq2)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	sessionID2 := listResp2.Header().Get("X-Session-ID")
	if sessionID2 == sessionID1 {
		t.Fatal("Expected different session ID")
	}

	if len(listResp2.Msg.Services) != 0 {
		t.Errorf("Expected zero services in new session, got %d", len(listResp2.Msg.Services))
	}
}
