package interfaces

import (
	"context"

	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ServiceRegistry provides methods for managing and querying service descriptors
type ServiceRegistry interface {
	// Register registers a file descriptor set
	Register(fds *descriptorpb.FileDescriptorSet) error

	// ListServices returns a list of all registered services
	ListServices() []ServiceInfo

	// GetServiceSchema returns detailed schema information for a service
	GetServiceSchema(serviceName string) (*ServiceInfo, map[string]string, error)

	// GetMethodDescriptor returns the method descriptor for a specific service and method
	GetMethodDescriptor(serviceName, methodName string) (*desc.MethodDescriptor, error)
}

// ServiceInfo contains information about a service
type ServiceInfo struct {
	Name    string
	Methods []MethodInfo
}

// MethodInfo contains information about a method
type MethodInfo struct {
	Name              string
	InputType         string
	OutputType        string
	ClientStreaming   bool
	ServerStreaming   bool
	InputMessageSchema  *MessageSchema
	OutputMessageSchema *MessageSchema
}

// MessageSchema describes the structure of a protobuf message
type MessageSchema struct {
	Name   string
	Fields []FieldInfo
}

// FieldInfo contains information about a message field
type FieldInfo struct {
	Name     string
	Type     string
	Number   int32
	Label    string
	Repeated bool
	Optional bool
	Required bool
}

// RpcInvoker provides methods for dynamically invoking RPC methods
type RpcInvoker interface {
	// InvokeUnary performs a unary RPC invocation
	InvokeUnary(ctx context.Context, req InvokeRequest) (*InvokeResponse, error)

	// Close closes all connections managed by the invoker
	Close() error

	// GetConnectionStats returns statistics about active connections
	GetConnectionStats() ConnectionStats
}

// InvokeRequest contains parameters for a dynamic RPC invocation
type InvokeRequest struct {
	Endpoint        string
	ServiceName     string
	MethodName      string
	RequestJSON     []byte
	UseTLS          bool
	ServerName      string
	TimeoutSeconds  int32
	Metadata        map[string]string
	MethodDesc      *desc.MethodDescriptor
	Transport       catalogv1.Transport
}

// InvokeResponse contains the result of an RPC invocation
type InvokeResponse struct {
	Success       bool
	ResponseJSON  []byte
	Error         string
	Metadata      map[string]string
	StatusCode    int32
	StatusMessage string
}

// ConnectionStats provides statistics about active connections
type ConnectionStats struct {
	TotalConnections  int
	ActiveConnections int
	EndpointCounts    map[string]int
}

// SessionManager manages session state and lifecycle
type SessionManager interface {
	// GetOrCreate retrieves an existing session or creates a new one
	GetOrCreate(sessionID string) (Session, string, error)

	// Get retrieves an existing session by ID
	Get(sessionID string) (Session, bool)

	// Delete removes a session by ID
	Delete(sessionID string)

	// Close closes all sessions and cleans up resources
	Close() error

	// Count returns the number of active sessions
	Count() int
}

// Session represents an active session with its state
type Session interface {
	// ID returns the unique session identifier
	ID() string

	// Registry returns the service registry for this session
	Registry() ServiceRegistry

	// Invoker returns the RPC invoker for this session
	Invoker() RpcInvoker

	// Close closes the session and cleans up resources
	Close() error
}

// ProtoLoader provides methods for loading protobuf definitions from various sources
type ProtoLoader interface {
	// LoadFromPath loads protobufs from a local file path
	LoadFromPath(path string) (*descriptorpb.FileDescriptorSet, error)

	// LoadFromGitHub loads protobufs from a GitHub repository
	LoadFromGitHub(repoURL string) (*descriptorpb.FileDescriptorSet, error)

	// LoadFromBufModule loads protobufs from a Buf Schema Registry module
	LoadFromBufModule(moduleRef string) (*descriptorpb.FileDescriptorSet, error)

	// LoadFromReflection loads protobufs using gRPC server reflection
	LoadFromReflection(endpoint string, useTLS bool, serverName string) (*descriptorpb.FileDescriptorSet, error)
}
