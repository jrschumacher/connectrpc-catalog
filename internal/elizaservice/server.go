package elizaservice

import (
	"context"
	"log"
	"net/http"

	"github.com/opentdf/connectrpc-catalog/gen/connectrpc/eliza/v1/elizav1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Server is an Eliza service server that supports Connect, gRPC, and gRPC-Web.
type Server struct {
	httpServer *http.Server
	port       string
}

// NewServer creates a new Eliza server on the specified port.
func NewServer(port string) *Server {
	mux := http.NewServeMux()

	// Create the Eliza service handler - this automatically supports
	// Connect, gRPC, and gRPC-Web protocols
	handler := NewHandler()
	path, elizaHandler := elizav1connect.NewElizaServiceHandler(handler)
	mux.Handle(path, elizaHandler)

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Wrap with h2c to support HTTP/2 cleartext (required for gRPC without TLS)
	h2cHandler := h2c.NewHandler(mux, &http2.Server{})

	return &Server{
		port: port,
		httpServer: &http.Server{
			Addr:    ":" + port,
			Handler: h2cHandler,
		},
	}
}

// Start starts the server (blocking).
func (s *Server) Start() error {
	log.Printf("Eliza service listening on port %s", s.port)
	log.Printf("Supported protocols: Connect (HTTP/JSON), gRPC (HTTP/2), gRPC-Web")
	log.Printf("Health check: http://localhost:%s/health", s.port)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
