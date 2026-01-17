package elizaservice_test

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	elizav1 "github.com/opentdf/connectrpc-catalog/gen/connectrpc/eliza/v1"
	"github.com/opentdf/connectrpc-catalog/gen/connectrpc/eliza/v1/elizav1connect"
	"github.com/opentdf/connectrpc-catalog/internal/elizaservice"
	"golang.org/x/net/http2"
)

func TestElizaService_AllProtocols(t *testing.T) {
	// Start the server
	server := elizaservice.NewServer("50099")
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

	baseURL := "http://localhost:50099"

	t.Run("Connect protocol", func(t *testing.T) {
		client := elizav1connect.NewElizaServiceClient(
			http.DefaultClient,
			baseURL,
			// Default is Connect protocol
		)

		req := connect.NewRequest(&elizav1.SayRequest{
			Sentence: "Hello from Connect test",
		})

		resp, err := client.Say(context.Background(), req)
		if err != nil {
			t.Fatalf("Connect Say failed: %v", err)
		}

		if resp.Msg.GetSentence() != "Hello! How can I help you today?" {
			t.Errorf("Unexpected response: %s", resp.Msg.GetSentence())
		}
		t.Logf("Connect response: %s", resp.Msg.GetSentence())
	})

	t.Run("gRPC protocol", func(t *testing.T) {
		// Create HTTP/2 client for gRPC (h2c - HTTP/2 cleartext)
		h2cClient := &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					// Dial without TLS for h2c
					return net.Dial(network, addr)
				},
			},
		}

		client := elizav1connect.NewElizaServiceClient(
			h2cClient,
			baseURL,
			connect.WithGRPC(), // Use gRPC protocol
		)

		req := connect.NewRequest(&elizav1.SayRequest{
			Sentence: "Hello from gRPC test",
		})

		resp, err := client.Say(context.Background(), req)
		if err != nil {
			t.Fatalf("gRPC Say failed: %v", err)
		}

		if resp.Msg.GetSentence() != "Hello! How can I help you today?" {
			t.Errorf("Unexpected response: %s", resp.Msg.GetSentence())
		}
		t.Logf("gRPC response: %s", resp.Msg.GetSentence())
	})

	t.Run("gRPC-Web protocol", func(t *testing.T) {
		client := elizav1connect.NewElizaServiceClient(
			http.DefaultClient,
			baseURL,
			connect.WithGRPCWeb(), // Use gRPC-Web protocol
		)

		req := connect.NewRequest(&elizav1.SayRequest{
			Sentence: "Hello from gRPC-Web test",
		})

		resp, err := client.Say(context.Background(), req)
		if err != nil {
			t.Fatalf("gRPC-Web Say failed: %v", err)
		}

		if resp.Msg.GetSentence() != "Hello! How can I help you today?" {
			t.Errorf("Unexpected response: %s", resp.Msg.GetSentence())
		}
		t.Logf("gRPC-Web response: %s", resp.Msg.GetSentence())
	})
}
