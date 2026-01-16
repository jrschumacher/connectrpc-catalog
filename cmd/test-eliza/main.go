// Command test-eliza starts an Eliza service for testing.
// The service supports all three protocols: Connect, gRPC, and gRPC-Web.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opentdf/connectrpc-catalog/internal/elizaservice"
)

func main() {
	port := flag.String("port", "50051", "Port to listen on")
	flag.Parse()

	server := elizaservice.NewServer(*port)

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down Eliza server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			log.Printf("Error shutting down: %v", err)
		}
	}()

	if err := server.Start(); err != nil {
		if err.Error() != "http: Server closed" {
			log.Fatalf("Server error: %v", err)
		}
	}
}
