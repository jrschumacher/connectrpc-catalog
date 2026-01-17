package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"connectrpc.com/connect"
	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
	catalogv1connect "github.com/opentdf/connectrpc-catalog/gen/catalog/v1/catalogv1connect"
	"github.com/opentdf/connectrpc-catalog/internal/server"
)

// TestIntegrationLoadProtos_LocalPath tests loading protos from a local filesystem path
func TestIntegrationLoadProtos_LocalPath(t *testing.T) {
	// Setup: Create test server
	catalogServer := server.New()
	defer catalogServer.Close()

	mux := http.NewServeMux()
	path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client := catalogv1connect.NewCatalogServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	// Get the path to test proto files
	protoPath := getTestProtoPath(t)

	// Test: Load protos from local path
	ctx := context.Background()
	req := connect.NewRequest(&catalogv1.LoadProtosRequest{
		Source: &catalogv1.LoadProtosRequest_ProtoPath{
			ProtoPath: protoPath,
		},
	})

	resp, err := client.LoadProtos(ctx, req)
	if err != nil {
		t.Fatalf("LoadProtos failed: %v", err)
	}

	// Verify: Response should indicate success
	if !resp.Msg.Success {
		t.Errorf("LoadProtos returned success=false, error: %s", resp.Msg.Error)
	}

	// Verify: Should have loaded at least 1 service and 1 file
	if resp.Msg.ServiceCount < 1 {
		t.Errorf("Expected at least 1 service, got %d", resp.Msg.ServiceCount)
	}
	if resp.Msg.FileCount < 1 {
		t.Errorf("Expected at least 1 file, got %d", resp.Msg.FileCount)
	}

	t.Logf("✅ Loaded %d services from %d files", resp.Msg.ServiceCount, resp.Msg.FileCount)
}

// TestIntegrationLoadProtos_InvalidPath tests error handling for invalid paths
func TestIntegrationLoadProtos_InvalidPath(t *testing.T) {
	catalogServer := server.New()
	defer catalogServer.Close()

	mux := http.NewServeMux()
	path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client := catalogv1connect.NewCatalogServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	// Test: Load protos from non-existent path
	ctx := context.Background()
	req := connect.NewRequest(&catalogv1.LoadProtosRequest{
		Source: &catalogv1.LoadProtosRequest_ProtoPath{
			ProtoPath: "/nonexistent/path/to/protos",
		},
	})

	resp, err := client.LoadProtos(ctx, req)
	if err != nil {
		t.Fatalf("LoadProtos request failed: %v", err)
	}

	// Verify: Response should indicate failure
	if resp.Msg.Success {
		t.Errorf("LoadProtos should fail for invalid path, but returned success=true")
	}

	// Verify: Error message should be present
	if resp.Msg.Error == "" {
		t.Errorf("Expected error message for invalid path, got empty string")
	}

	t.Logf("✅ Correctly handled invalid path with error: %s", resp.Msg.Error)
}

// TestIntegrationListServices tests listing loaded services
func TestIntegrationListServices(t *testing.T) {
	// Setup: Create and configure server
	catalogServer := server.New()
	defer catalogServer.Close()

	mux := http.NewServeMux()
	path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client := catalogv1connect.NewCatalogServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	ctx := context.Background()

	// Step 1: Load protos
	protoPath := getTestProtoPath(t)
	loadReq := connect.NewRequest(&catalogv1.LoadProtosRequest{
		Source: &catalogv1.LoadProtosRequest_ProtoPath{
			ProtoPath: protoPath,
		},
	})

	loadResp, err := client.LoadProtos(ctx, loadReq)
	if err != nil {
		t.Fatalf("LoadProtos failed: %v", err)
	}
	if !loadResp.Msg.Success {
		t.Fatalf("LoadProtos returned error: %s", loadResp.Msg.Error)
	}

	// Get session ID from response
	sessionID := loadResp.Header().Get("X-Session-ID")
	if sessionID == "" {
		t.Fatal("Expected X-Session-ID header in LoadProtos response")
	}

	// Step 2: List services using the same session
	listReq := connect.NewRequest(&catalogv1.ListServicesRequest{})
	listReq.Header().Set("X-Session-ID", sessionID)
	listResp, err := client.ListServices(ctx, listReq)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	// Verify: Should have services
	if len(listResp.Msg.Services) == 0 {
		t.Errorf("Expected at least 1 service, got 0")
	}

	// Verify: Each service should have required fields
	for i, svc := range listResp.Msg.Services {
		if svc.Name == "" {
			t.Errorf("Service %d has empty name", i)
		}
		if len(svc.Methods) == 0 {
			t.Errorf("Service %s has no methods", svc.Name)
		}

		// Verify: Each method should have required fields
		for j, method := range svc.Methods {
			if method.Name == "" {
				t.Errorf("Service %s method %d has empty name", svc.Name, j)
			}
			if method.InputType == "" {
				t.Errorf("Service %s method %s has empty input type", svc.Name, method.Name)
			}
			if method.OutputType == "" {
				t.Errorf("Service %s method %s has empty output type", svc.Name, method.Name)
			}
		}

		t.Logf("✅ Service: %s (%d methods)", svc.Name, len(svc.Methods))
	}
}

// TestIntegrationGetServiceSchema tests retrieving service schema
func TestIntegrationGetServiceSchema(t *testing.T) {
	// Setup: Create and configure server
	catalogServer := server.New()
	defer catalogServer.Close()

	mux := http.NewServeMux()
	path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client := catalogv1connect.NewCatalogServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	ctx := context.Background()

	// Step 1: Load protos
	protoPath := getTestProtoPath(t)
	loadReq := connect.NewRequest(&catalogv1.LoadProtosRequest{
		Source: &catalogv1.LoadProtosRequest_ProtoPath{
			ProtoPath: protoPath,
		},
	})

	loadResp, err := client.LoadProtos(ctx, loadReq)
	if err != nil {
		t.Fatalf("LoadProtos failed: %v", err)
	}
	if !loadResp.Msg.Success {
		t.Fatalf("LoadProtos returned error: %s", loadResp.Msg.Error)
	}

	// Get session ID from response
	sessionID := loadResp.Header().Get("X-Session-ID")
	if sessionID == "" {
		t.Fatal("Expected X-Session-ID header in LoadProtos response")
	}

	// Step 2: Get first service name using the same session
	listReq := connect.NewRequest(&catalogv1.ListServicesRequest{})
	listReq.Header().Set("X-Session-ID", sessionID)
	listResp, err := client.ListServices(ctx, listReq)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}
	if len(listResp.Msg.Services) == 0 {
		t.Fatal("No services available to test GetServiceSchema")
	}

	serviceName := listResp.Msg.Services[0].Name

	// Step 3: Get service schema using the same session
	schemaReq := connect.NewRequest(&catalogv1.GetServiceSchemaRequest{
		ServiceName: serviceName,
	})
	schemaReq.Header().Set("X-Session-ID", sessionID)

	schemaResp, err := client.GetServiceSchema(ctx, schemaReq)
	if err != nil {
		t.Fatalf("GetServiceSchema failed: %v", err)
	}

	// Verify: Should have service info
	if schemaResp.Msg.Service == nil {
		t.Errorf("Expected service info, got nil")
	}

	// Verify: Should have message schemas
	if len(schemaResp.Msg.MessageSchemas) == 0 {
		t.Errorf("Expected message schemas, got empty map")
	}

	// Verify: Error field should be empty on success
	if schemaResp.Msg.Error != "" {
		t.Errorf("Unexpected error in response: %s", schemaResp.Msg.Error)
	}

	t.Logf("✅ Retrieved schema for service: %s", serviceName)
	t.Logf("   - Methods: %d", len(schemaResp.Msg.Service.Methods))
	t.Logf("   - Message schemas: %d", len(schemaResp.Msg.MessageSchemas))
}

// TestIntegrationGetServiceSchema_InvalidService tests error handling for invalid service
func TestIntegrationGetServiceSchema_InvalidService(t *testing.T) {
	catalogServer := server.New()
	defer catalogServer.Close()

	mux := http.NewServeMux()
	path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client := catalogv1connect.NewCatalogServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	ctx := context.Background()

	// Test: Get schema for non-existent service
	schemaReq := connect.NewRequest(&catalogv1.GetServiceSchemaRequest{
		ServiceName: "nonexistent.v1.FakeService",
	})

	schemaResp, err := client.GetServiceSchema(ctx, schemaReq)
	if err != nil {
		t.Fatalf("GetServiceSchema request failed: %v", err)
	}

	// Verify: Error field should contain error message
	if schemaResp.Msg.Error == "" {
		t.Errorf("Expected error for non-existent service, got empty error")
	}

	t.Logf("✅ Correctly handled invalid service with error: %s", schemaResp.Msg.Error)
}

// TestIntegrationListServices_EmptyRegistry tests listing services when registry is empty
func TestIntegrationListServices_EmptyRegistry(t *testing.T) {
	catalogServer := server.New()
	defer catalogServer.Close()

	mux := http.NewServeMux()
	path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client := catalogv1connect.NewCatalogServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	ctx := context.Background()

	// Test: List services from empty registry
	listReq := connect.NewRequest(&catalogv1.ListServicesRequest{})
	listResp, err := client.ListServices(ctx, listReq)
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	// Verify: Should return empty list or nil (both acceptable)
	if listResp.Msg.Services != nil && len(listResp.Msg.Services) != 0 {
		t.Errorf("Expected 0 services in empty registry, got %d", len(listResp.Msg.Services))
	}

	serviceCount := 0
	if listResp.Msg.Services != nil {
		serviceCount = len(listResp.Msg.Services)
	}

	t.Logf("✅ Empty registry correctly returns %d services", serviceCount)
}

// TestIntegrationInvokeGRPC_MissingFields tests validation of required fields
func TestIntegrationInvokeGRPC_MissingFields(t *testing.T) {
	catalogServer := server.New()
	defer catalogServer.Close()

	mux := http.NewServeMux()
	path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client := catalogv1connect.NewCatalogServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	ctx := context.Background()

	tests := []struct {
		name        string
		req         *catalogv1.InvokeGRPCRequest
		expectError bool
	}{
		{
			name: "missing endpoint",
			req: &catalogv1.InvokeGRPCRequest{
				Service:     "test.Service",
				Method:      "TestMethod",
				RequestJson: "{}",
			},
			expectError: true,
		},
		{
			name: "missing service",
			req: &catalogv1.InvokeGRPCRequest{
				Endpoint:    "localhost:8080",
				Method:      "TestMethod",
				RequestJson: "{}",
			},
			expectError: true,
		},
		{
			name: "missing method",
			req: &catalogv1.InvokeGRPCRequest{
				Endpoint:    "localhost:8080",
				Service:     "test.Service",
				RequestJson: "{}",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invokeReq := connect.NewRequest(tt.req)
			_, err := client.InvokeGRPC(ctx, invokeReq)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}

			if tt.expectError && err != nil {
				t.Logf("✅ Correctly rejected request: %v", err)
			}
		})
	}
}

// TestIntegrationMultipleLoadProtos tests loading protos multiple times
func TestIntegrationMultipleLoadProtos(t *testing.T) {
	catalogServer := server.New()
	defer catalogServer.Close()

	mux := http.NewServeMux()
	path, handler := catalogv1connect.NewCatalogServiceHandler(catalogServer)
	mux.Handle(path, handler)

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	client := catalogv1connect.NewCatalogServiceClient(
		http.DefaultClient,
		testServer.URL,
	)

	ctx := context.Background()
	protoPath := getTestProtoPath(t)

	// Load protos first time
	loadReq1 := connect.NewRequest(&catalogv1.LoadProtosRequest{
		Source: &catalogv1.LoadProtosRequest_ProtoPath{
			ProtoPath: protoPath,
		},
	})

	resp1, err := client.LoadProtos(ctx, loadReq1)
	if err != nil {
		t.Fatalf("First LoadProtos failed: %v", err)
	}
	if !resp1.Msg.Success {
		t.Fatalf("First LoadProtos returned error: %s", resp1.Msg.Error)
	}

	initialServiceCount := resp1.Msg.ServiceCount

	// Load protos second time (should add to existing)
	loadReq2 := connect.NewRequest(&catalogv1.LoadProtosRequest{
		Source: &catalogv1.LoadProtosRequest_ProtoPath{
			ProtoPath: protoPath,
		},
	})

	resp2, err := client.LoadProtos(ctx, loadReq2)
	if err != nil {
		t.Fatalf("Second LoadProtos failed: %v", err)
	}
	if !resp2.Msg.Success {
		t.Fatalf("Second LoadProtos returned error: %s", resp2.Msg.Error)
	}

	// Verify: Can load multiple times without error
	t.Logf("✅ Successfully loaded protos twice")
	t.Logf("   - First load: %d services", initialServiceCount)
	t.Logf("   - Second load: %d services", resp2.Msg.ServiceCount)
}

// Helper function to get test proto path
func getTestProtoPath(t *testing.T) string {
	t.Helper()

	// Try to find proto directory relative to test location
	candidates := []string{
		"../../proto",                                          // From internal/server
		"./proto",                                              // From project root
		"../proto",                                             // From internal
		filepath.Join(os.Getenv("PWD"), "proto"),              // Using PWD
		"/Users/jschumacher/Projects/connectrpc-catalog/proto", // Absolute fallback
	}

	for _, path := range candidates {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		if stat, err := os.Stat(absPath); err == nil && stat.IsDir() {
			return absPath
		}
	}

	t.Fatal("Could not find proto directory. Tried: " + filepath.Join(candidates...))
	return ""
}
