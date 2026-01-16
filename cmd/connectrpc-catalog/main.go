package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/connect"
	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
	catalogv1connect "github.com/opentdf/connectrpc-catalog/gen/catalog/v1/catalogv1connect"
	"github.com/opentdf/connectrpc-catalog/internal/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

//go:embed all:dist
var uiAssets embed.FS

const (
	defaultPort = "8080"
	defaultHost = "localhost"
)

func main() {
	// Parse command-line flags
	var (
		port         = flag.String("port", defaultPort, "HTTP server port")
		host         = flag.String("host", defaultHost, "HTTP server host")
		protoPath    = flag.String("proto-path", "", "Local directory path for proto files")
		protoRepo    = flag.String("proto-repo", "", "GitHub repository (e.g., github.com/connectrpc/eliza)")
		bufModule    = flag.String("buf-module", "", "Buf registry module (e.g., buf.build/connectrpc/eliza)")
		endpoint     = flag.String("endpoint", "", "Default gRPC endpoint for invocations (optional)")
	)
	flag.Parse()

	// Create catalog server
	catalogServer := server.New()
	defer func() {
		if err := catalogServer.Close(); err != nil {
			log.Printf("Error closing catalog server: %v", err)
		}
	}()

	// Validate server setup
	if err := catalogServer.ValidateSetup(); err != nil {
		log.Fatalf("Server setup validation failed: %v", err)
	}

	// Auto-load protos if source flags are provided
	if err := loadProtosFromFlags(catalogServer, *protoPath, *protoRepo, *bufModule, *endpoint); err != nil {
		log.Printf("Warning: Failed to auto-load protos: %v", err)
		// Continue server startup even if proto loading fails
	}

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register Connect handlers with CORS wrapper
	path, handler := catalogv1connect.NewCatalogServiceHandler(
		catalogServer,
		connect.WithInterceptors(corsInterceptor()),
	)
	// Wrap handler with CORS middleware for preflight requests
	mux.Handle(path, corsMiddleware(handler))

	// Serve embedded UI assets
	uiFS, err := fs.Sub(uiAssets, "dist")
	if err != nil {
		log.Fatalf("Failed to get UI filesystem: %v", err)
	}

	// Register MIME types for common web assets
	registerMIMETypes()

	// Serve static files with SPA fallback
	mux.HandleFunc("/", spaHandler(uiFS))

	// Create server with h2c support (HTTP/2 without TLS) for Connect
	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", *host, *port),
		Handler: h2c.NewHandler(mux, h2s),
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Printf("ConnectRPC Catalog server starting on http://%s:%s", *host, *port)
		log.Printf("UI available at: http://%s:%s", *host, *port)
		log.Printf("API available at: http://%s:%s/catalog.v1.CatalogService/*", *host, *port)

		if err := h1s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Shutting down server gracefully...")

	// Close server
	if err := h1s.Close(); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// spaHandler serves static files and falls back to index.html for client-side routing
func spaHandler(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Don't handle API routes
		if strings.HasPrefix(r.URL.Path, "/catalog.v1.CatalogService/") {
			http.NotFound(w, r)
			return
		}

		// Clean the path
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try to open the requested file
		file, err := fsys.Open(path)
		if err != nil {
			// File not found, serve index.html for SPA routing
			indexFile, indexErr := fsys.Open("index.html")
			if indexErr != nil {
				http.Error(w, "index.html not found", http.StatusInternalServerError)
				return
			}
			defer indexFile.Close()

			// Set correct content type for HTML
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(w, r, "index.html", getModTime(indexFile), indexFile.(io.ReadSeeker))
			return
		}
		defer file.Close()

		// Determine content type based on file extension
		ext := filepath.Ext(path)
		if contentType := mime.TypeByExtension(ext); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		// Serve the file
		http.ServeContent(w, r, path, getModTime(file), file.(io.ReadSeeker))
	}
}

// getModTime extracts modification time from file info
func getModTime(file fs.File) time.Time {
	if stat, err := file.Stat(); err == nil {
		return stat.ModTime()
	}
	return time.Time{}
}

// registerMIMETypes ensures proper MIME types for web assets
func registerMIMETypes() {
	mimeTypes := map[string]string{
		".js":   "application/javascript",
		".mjs":  "application/javascript",
		".json": "application/json",
		".css":  "text/css",
		".html": "text/html; charset=utf-8",
		".svg":  "image/svg+xml",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".woff": "font/woff",
		".woff2": "font/woff2",
		".ttf":  "font/ttf",
		".eot":  "application/vnd.ms-fontobject",
		".ico":  "image/x-icon",
	}

	for ext, mimeType := range mimeTypes {
		if err := mime.AddExtensionType(ext, mimeType); err != nil {
			// Type might already be registered, which is fine
			log.Printf("Note: could not register MIME type for %s: %v", ext, err)
		}
	}
}

// corsMiddleware wraps an http.Handler to add CORS headers and handle preflight requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			setCORSHeaders(w)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set CORS headers for all requests
		setCORSHeaders(w)
		next.ServeHTTP(w, r)
	})
}

// setCORSHeaders sets CORS headers for Connect requests
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// corsInterceptor creates a Connect interceptor that adds CORS headers
func corsInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Note: CORS headers are set in the HTTP handler wrapper
			return next(ctx, req)
		}
	}
}

// loadProtosFromFlags handles auto-loading protos from CLI flags
func loadProtosFromFlags(catalogServer *server.CatalogServer, protoPath, protoRepo, bufModule, endpoint string) error {
	// Count how many proto sources are provided
	sourcesProvided := 0
	if protoPath != "" {
		sourcesProvided++
	}
	if protoRepo != "" {
		sourcesProvided++
	}
	if bufModule != "" {
		sourcesProvided++
	}

	// No source provided - nothing to do
	if sourcesProvided == 0 {
		return nil
	}

	// Validate that only ONE source is provided
	if sourcesProvided > 1 {
		return fmt.Errorf("only one proto source flag can be specified at a time (--proto-path, --proto-repo, or --buf-module)")
	}

	// Build the LoadProtos request based on which flag was provided
	var req *connect.Request[catalogv1.LoadProtosRequest]

	switch {
	case protoPath != "":
		log.Printf("Auto-loading protos from local path: %s", protoPath)
		req = connect.NewRequest(&catalogv1.LoadProtosRequest{
			Source: &catalogv1.LoadProtosRequest_ProtoPath{
				ProtoPath: protoPath,
			},
		})

	case protoRepo != "":
		log.Printf("Auto-loading protos from GitHub repository: %s", protoRepo)
		req = connect.NewRequest(&catalogv1.LoadProtosRequest{
			Source: &catalogv1.LoadProtosRequest_ProtoRepo{
				ProtoRepo: protoRepo,
			},
		})

	case bufModule != "":
		log.Printf("Auto-loading protos from Buf module: %s", bufModule)
		req = connect.NewRequest(&catalogv1.LoadProtosRequest{
			Source: &catalogv1.LoadProtosRequest_BufModule{
				BufModule: bufModule,
			},
		})
	}

	// Call LoadProtos
	ctx := context.Background()
	resp, err := catalogServer.LoadProtos(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to call LoadProtos: %w", err)
	}

	// Check response
	if !resp.Msg.Success {
		return fmt.Errorf("proto loading failed: %s", resp.Msg.Error)
	}

	log.Printf("Successfully loaded protos: %d services from %d files", resp.Msg.ServiceCount, resp.Msg.FileCount)

	// Log endpoint configuration if provided
	if endpoint != "" {
		log.Printf("Default endpoint configured: %s (can be changed in UI)", endpoint)
		// Note: Endpoint is stored in UI localStorage, not server-side
		// This is just informational for the user
	}

	return nil
}
