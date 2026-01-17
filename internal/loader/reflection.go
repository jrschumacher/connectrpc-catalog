// Package loader provides functionality for loading protocol buffer descriptors
// from various sources including gRPC server reflection.
//
// The reflection loader connects to a gRPC server that supports the server
// reflection protocol and retrieves service descriptors dynamically at runtime.
// This is useful for service discovery and introspection without needing
// access to .proto files.
package loader

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ReflectionOptions configures reflection-based discovery
type ReflectionOptions struct {
	UseTLS         bool
	ServerName     string
	TimeoutSeconds int32
}

// LoadFromReflection fetches proto descriptors from a gRPC server via reflection
func LoadFromReflection(endpoint string, opts ReflectionOptions) (*descriptorpb.FileDescriptorSet, error) {
	// Set default timeout
	timeout := time.Duration(opts.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Configure dial options
	var dialOpts []grpc.DialOption
	if opts.UseTLS {
		tlsConfig := &tls.Config{}
		if opts.ServerName != "" {
			tlsConfig.ServerName = opts.ServerName
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Connect to the server
	conn, err := grpc.DialContext(ctx, endpoint, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", endpoint, err)
	}
	defer conn.Close()

	// Create reflection client (try v1alpha first, most common)
	refClient := grpcreflect.NewClientV1Alpha(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(conn))
	defer refClient.Reset()

	// List all services
	services, err := refClient.ListServices()
	if err != nil {
		return nil, fmt.Errorf("failed to list services via reflection: %w", err)
	}

	// Collect all file descriptors
	fileDescriptors := make(map[string]*desc.FileDescriptor)

	for _, svcName := range services {
		// Skip reflection service itself
		if svcName == "grpc.reflection.v1alpha.ServerReflection" ||
			svcName == "grpc.reflection.v1.ServerReflection" {
			continue
		}

		// Get file descriptor for this service
		fd, err := refClient.FileContainingSymbol(svcName)
		if err != nil {
			// Log warning but continue with other services
			fmt.Printf("Warning: could not get descriptor for %s: %v\n", svcName, err)
			continue
		}

		// Collect this file and all its dependencies
		collectFileDescriptors(fd, fileDescriptors)
	}

	if len(fileDescriptors) == 0 {
		return nil, fmt.Errorf("no service descriptors found via reflection")
	}

	// Convert to FileDescriptorSet
	fds := &descriptorpb.FileDescriptorSet{
		File: make([]*descriptorpb.FileDescriptorProto, 0, len(fileDescriptors)),
	}

	for _, fd := range fileDescriptors {
		fds.File = append(fds.File, fd.AsFileDescriptorProto())
	}

	return fds, nil
}

// collectFileDescriptors recursively collects a file descriptor and all its dependencies
func collectFileDescriptors(fd *desc.FileDescriptor, collected map[string]*desc.FileDescriptor) {
	name := fd.GetName()
	if _, exists := collected[name]; exists {
		return
	}
	collected[name] = fd

	// Collect dependencies
	for _, dep := range fd.GetDependencies() {
		collectFileDescriptors(dep, collected)
	}
}

// CheckReflectionSupport tests if an endpoint supports gRPC reflection
func CheckReflectionSupport(endpoint string, useTLS bool) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var dialOpts []grpc.DialOption
	if useTLS {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, endpoint, dialOpts...)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	refClient := grpcreflect.NewClientV1Alpha(ctx, grpc_reflection_v1alpha.NewServerReflectionClient(conn))
	defer refClient.Reset()

	_, err = refClient.ListServices()
	return err == nil, err
}
