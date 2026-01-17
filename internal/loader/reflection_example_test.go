package loader_test

import (
	"fmt"

	"github.com/opentdf/connectrpc-catalog/internal/loader"
)

// ExampleLoadFromReflection demonstrates loading proto descriptors from a gRPC server
func ExampleLoadFromReflection() {
	// Configure reflection options
	opts := loader.ReflectionOptions{
		UseTLS:         false, // Set to true for TLS connections
		ServerName:     "",    // Optional: specify server name for TLS
		TimeoutSeconds: 10,    // Connection timeout
	}

	// Load descriptors from reflection-enabled gRPC server
	fds, err := loader.LoadFromReflection("localhost:50051", opts)
	if err != nil {
		// Handle error (connection failed, no reflection support, etc.)
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Get information about loaded descriptors
	info := loader.GetDescriptorInfo(fds)
	fmt.Printf("Loaded %d files with %d services\n", info.Files, len(info.Services))

	// Print service names
	for _, svc := range info.Services {
		fmt.Printf("Service: %s\n", svc)
	}
}

// ExampleCheckReflectionSupport demonstrates checking if a server supports reflection
func ExampleCheckReflectionSupport() {
	// Check if server supports gRPC reflection
	supported, err := loader.CheckReflectionSupport("localhost:50051", false)
	if err != nil {
		fmt.Printf("Error checking reflection support: %v\n", err)
		return
	}

	if supported {
		fmt.Println("Server supports gRPC reflection")
	} else {
		fmt.Println("Server does not support gRPC reflection")
	}
}

// ExampleLoad_reflection demonstrates using the unified Load function with reflection
func ExampleLoad_reflection() {
	// Create a reflection source
	source := loader.LoadSource{
		Type:  loader.SourceTypeReflection,
		Value: "localhost:50051",
		ReflectionOptions: &loader.ReflectionOptions{
			UseTLS:         false,
			TimeoutSeconds: 10,
		},
	}

	// Load using unified loader
	fds, err := loader.Load(source)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Get information about loaded descriptors
	info := loader.GetDescriptorInfo(fds)
	fmt.Printf("Loaded %d services\n", len(info.Services))
}

// ExampleReflectionOptions demonstrates various configuration options
func ExampleReflectionOptions() {
	// Basic configuration (no TLS)
	basicOpts := loader.ReflectionOptions{
		UseTLS:         false,
		TimeoutSeconds: 10,
	}
	fmt.Printf("Basic: TLS=%v, Timeout=%ds\n", basicOpts.UseTLS, basicOpts.TimeoutSeconds)

	// TLS configuration
	tlsOpts := loader.ReflectionOptions{
		UseTLS:         true,
		ServerName:     "api.example.com",
		TimeoutSeconds: 30,
	}
	fmt.Printf("TLS: TLS=%v, ServerName=%s, Timeout=%ds\n",
		tlsOpts.UseTLS, tlsOpts.ServerName, tlsOpts.TimeoutSeconds)

	// Default timeout (will use 10 seconds)
	defaultOpts := loader.ReflectionOptions{
		UseTLS: false,
	}
	fmt.Printf("Default: TLS=%v, Timeout=%ds (uses default)\n",
		defaultOpts.UseTLS, defaultOpts.TimeoutSeconds)

	// Output:
	// Basic: TLS=false, Timeout=10s
	// TLS: TLS=true, ServerName=api.example.com, Timeout=30s
	// Default: TLS=false, Timeout=0s (uses default)
}
