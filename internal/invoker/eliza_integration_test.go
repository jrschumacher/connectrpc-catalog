package invoker_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
	"github.com/opentdf/connectrpc-catalog/internal/elizaservice"
	"github.com/opentdf/connectrpc-catalog/internal/invoker"
	"github.com/opentdf/connectrpc-catalog/internal/loader"
	"github.com/opentdf/connectrpc-catalog/internal/registry"
)

func TestInvoker_ElizaIntegration(t *testing.T) {
	// Start the Eliza server
	server := elizaservice.NewServer("50097")
	go func() {
		if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
			t.Logf("Server error: %v", err)
		}
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Load Eliza protos from BSR
	fds, err := loader.LoadFromBufModule("buf.build/connectrpc/eliza")
	if err != nil {
		t.Fatalf("Failed to load Eliza protos: %v", err)
	}

	// Create registry and register the descriptors
	reg := registry.New()
	if err := reg.Register(fds); err != nil {
		t.Fatalf("Failed to register descriptors: %v", err)
	}

	// Get the Say method descriptor
	sayMethodDesc, err := reg.GetMethodDescriptor("connectrpc.eliza.v1.ElizaService", "Say")
	if err != nil {
		t.Fatalf("Could not find Say method: %v", err)
	}

	inv := invoker.New()
	defer inv.Close()

	t.Run("Connect protocol", func(t *testing.T) {
		req := invoker.InvokeRequest{
			Endpoint:       "localhost:50097",
			ServiceName:    "connectrpc.eliza.v1.ElizaService",
			MethodName:     "Say",
			RequestJSON:    json.RawMessage(`{"sentence": "Hello from Connect"}`),
			UseTLS:         false,
			TimeoutSeconds: 30,
			MethodDesc:     sayMethodDesc,
			Transport:      catalogv1.Transport_TRANSPORT_CONNECT,
		}

		resp, err := inv.InvokeUnary(context.Background(), req)
		if err != nil {
			t.Fatalf("Connect invocation error: %v", err)
		}

		if !resp.Success {
			t.Fatalf("Connect invocation failed: %s", resp.Error)
		}

		t.Logf("Connect response: %s", resp.ResponseJSON)
	})

	t.Run("gRPC protocol", func(t *testing.T) {
		req := invoker.InvokeRequest{
			Endpoint:       "localhost:50097",
			ServiceName:    "connectrpc.eliza.v1.ElizaService",
			MethodName:     "Say",
			RequestJSON:    json.RawMessage(`{"sentence": "Hello from gRPC"}`),
			UseTLS:         false,
			TimeoutSeconds: 30,
			MethodDesc:     sayMethodDesc,
			Transport:      catalogv1.Transport_TRANSPORT_GRPC,
		}

		resp, err := inv.InvokeUnary(context.Background(), req)
		if err != nil {
			t.Fatalf("gRPC invocation error: %v", err)
		}

		if !resp.Success {
			t.Fatalf("gRPC invocation failed: %s", resp.Error)
		}

		t.Logf("gRPC response: %s", resp.ResponseJSON)
	})
}
