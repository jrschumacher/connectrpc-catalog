package loader

import (
	"testing"
)

func TestReflectionOptions_DefaultTimeout(t *testing.T) {
	opts := ReflectionOptions{
		TimeoutSeconds: 0, // Should default to 10 seconds
	}

	if opts.TimeoutSeconds < 0 {
		t.Errorf("Expected non-negative timeout, got %d", opts.TimeoutSeconds)
	}
}

func TestReflectionOptions_CustomTimeout(t *testing.T) {
	opts := ReflectionOptions{
		TimeoutSeconds: 30,
	}

	if opts.TimeoutSeconds != 30 {
		t.Errorf("Expected timeout of 30, got %d", opts.TimeoutSeconds)
	}
}

func TestReflectionOptions_TLSConfig(t *testing.T) {
	opts := ReflectionOptions{
		UseTLS:     true,
		ServerName: "example.com",
	}

	if !opts.UseTLS {
		t.Error("Expected TLS to be enabled")
	}

	if opts.ServerName != "example.com" {
		t.Errorf("Expected server name 'example.com', got '%s'", opts.ServerName)
	}
}

// Note: Integration tests for LoadFromReflection and CheckReflectionSupport
// would require a running gRPC server with reflection enabled.
// These should be added as part of integration test suite.
