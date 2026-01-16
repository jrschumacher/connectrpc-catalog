package registry

import (
	"fmt"
	"sync"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Registry maintains an in-memory descriptor registry for dynamic gRPC invocation
type Registry struct {
	mu       sync.RWMutex
	files    map[string]*desc.FileDescriptor
	services map[string]*desc.ServiceDescriptor
	messages map[string]*desc.MessageDescriptor
}

// New creates a new empty registry
func New() *Registry {
	return &Registry{
		files:    make(map[string]*desc.FileDescriptor),
		services: make(map[string]*desc.ServiceDescriptor),
		messages: make(map[string]*desc.MessageDescriptor),
	}
}

// Register adds a FileDescriptorSet to the registry
func (r *Registry) Register(fds *descriptorpb.FileDescriptorSet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Convert descriptorpb to protoreflect FileDescriptor
	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return fmt.Errorf("failed to create file registry: %w", err)
	}

	// Process each file descriptor
	for _, fdpb := range fds.File {
		// Convert to jhump/protoreflect descriptor for easier access
		fd, err := desc.CreateFileDescriptor(fdpb)
		if err != nil {
			return fmt.Errorf("failed to create file descriptor for %s: %w", fdpb.GetName(), err)
		}

		// Store file descriptor
		r.files[fd.GetName()] = fd

		// Index services
		for _, svc := range fd.GetServices() {
			r.services[svc.GetFullyQualifiedName()] = svc
		}

		// Index messages
		for _, msg := range fd.GetMessageTypes() {
			r.indexMessage(msg)
		}
	}

	// Also process using protoreflect for additional validation
	var processErr error
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		// Additional processing if needed
		return true
	})

	if processErr != nil {
		return processErr
	}

	return nil
}

// indexMessage recursively indexes a message and its nested types
func (r *Registry) indexMessage(msg *desc.MessageDescriptor) {
	r.messages[msg.GetFullyQualifiedName()] = msg

	// Index nested messages
	for _, nested := range msg.GetNestedMessageTypes() {
		r.indexMessage(nested)
	}
}

// ServiceInfo contains metadata about a gRPC service
type ServiceInfo struct {
	Name          string
	Package       string
	Methods       []MethodInfo
	Documentation string
}

// MethodInfo contains metadata about a gRPC method
type MethodInfo struct {
	Name            string
	InputType       string
	OutputType      string
	Documentation   string
	ClientStreaming bool
	ServerStreaming bool
}

// ListServices returns all registered services
func (r *Registry) ListServices() []ServiceInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]ServiceInfo, 0, len(r.services))
	for _, svc := range r.services {
		info := ServiceInfo{
			Name:          svc.GetFullyQualifiedName(),
			Package:       svc.GetFile().GetPackage(),
			Documentation: extractComments(svc.GetSourceInfo()),
			Methods:       make([]MethodInfo, 0, len(svc.GetMethods())),
		}

		for _, method := range svc.GetMethods() {
			methodInfo := MethodInfo{
				Name:            method.GetName(),
				InputType:       method.GetInputType().GetFullyQualifiedName(),
				OutputType:      method.GetOutputType().GetFullyQualifiedName(),
				Documentation:   extractComments(method.GetSourceInfo()),
				ClientStreaming: method.IsClientStreaming(),
				ServerStreaming: method.IsServerStreaming(),
			}
			info.Methods = append(info.Methods, methodInfo)
		}

		services = append(services, info)
	}

	return services
}

// GetService retrieves a service descriptor by fully qualified name
func (r *Registry) GetService(name string) (*desc.ServiceDescriptor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", name)
	}

	return svc, nil
}

// GetMethodDescriptor retrieves a method descriptor by service and method name
func (r *Registry) GetMethodDescriptor(serviceName, methodName string) (*desc.MethodDescriptor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, exists := r.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	method := svc.FindMethodByName(methodName)
	if method == nil {
		return nil, fmt.Errorf("method not found: %s.%s", serviceName, methodName)
	}

	return method, nil
}

// GetMessageDescriptor retrieves a message descriptor by fully qualified name
func (r *Registry) GetMessageDescriptor(msgName string) (*desc.MessageDescriptor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, exists := r.messages[msgName]
	if !exists {
		return nil, fmt.Errorf("message not found: %s", msgName)
	}

	return msg, nil
}

// GetServiceSchema returns detailed schema information for a service
func (r *Registry) GetServiceSchema(serviceName string) (*ServiceInfo, map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, exists := r.services[serviceName]
	if !exists {
		return nil, nil, fmt.Errorf("service not found: %s", serviceName)
	}

	// Build service info
	info := ServiceInfo{
		Name:          svc.GetFullyQualifiedName(),
		Package:       svc.GetFile().GetPackage(),
		Documentation: extractComments(svc.GetSourceInfo()),
		Methods:       make([]MethodInfo, 0, len(svc.GetMethods())),
	}

	// Track all message types used by this service
	messageSchemas := make(map[string]string)
	messagesSeen := make(map[string]bool)

	for _, method := range svc.GetMethods() {
		methodInfo := MethodInfo{
			Name:            method.GetName(),
			InputType:       method.GetInputType().GetFullyQualifiedName(),
			OutputType:      method.GetOutputType().GetFullyQualifiedName(),
			Documentation:   extractComments(method.GetSourceInfo()),
			ClientStreaming: method.IsClientStreaming(),
			ServerStreaming: method.IsServerStreaming(),
		}
		info.Methods = append(info.Methods, methodInfo)

		// Collect schemas for input and output types
		r.collectMessageSchema(method.GetInputType(), messageSchemas, messagesSeen)
		r.collectMessageSchema(method.GetOutputType(), messageSchemas, messagesSeen)
	}

	return &info, messageSchemas, nil
}

// collectMessageSchema recursively collects JSON Schema for a message and its dependencies
func (r *Registry) collectMessageSchema(msg *desc.MessageDescriptor, schemas map[string]string, seen map[string]bool) {
	name := msg.GetFullyQualifiedName()
	if seen[name] {
		return
	}
	seen[name] = true

	// Generate JSON Schema for this message
	schema := r.generateJSONSchema(msg)
	schemas[name] = schema

	// Recursively process field types
	for _, field := range msg.GetFields() {
		if field.GetMessageType() != nil {
			r.collectMessageSchema(field.GetMessageType(), schemas, seen)
		}
	}

	// Process nested types
	for _, nested := range msg.GetNestedMessageTypes() {
		r.collectMessageSchema(nested, schemas, seen)
	}
}

// generateJSONSchema generates a JSON Schema representation of a message
func (r *Registry) generateJSONSchema(msg *desc.MessageDescriptor) string {
	// Simplified JSON Schema generation
	// In production, use a proper JSON Schema generator
	schema := fmt.Sprintf(`{
  "type": "object",
  "title": "%s",
  "properties": {`, msg.GetName())

	for i, field := range msg.GetFields() {
		if i > 0 {
			schema += ","
		}
		fieldType := getJSONType(field)
		schema += fmt.Sprintf(`
    "%s": {
      "type": "%s"`, field.GetName(), fieldType)

		if field.GetMessageType() != nil {
			schema += fmt.Sprintf(`,
      "$ref": "#/definitions/%s"`, field.GetMessageType().GetFullyQualifiedName())
		}

		schema += `
    }`
	}

	schema += `
  }
}`

	return schema
}

// getJSONType maps protobuf field types to JSON types
func getJSONType(field *desc.FieldDescriptor) string {
	switch field.GetType().String() {
	case "TYPE_DOUBLE", "TYPE_FLOAT":
		return "number"
	case "TYPE_INT32", "TYPE_INT64", "TYPE_UINT32", "TYPE_UINT64",
		"TYPE_SINT32", "TYPE_SINT64", "TYPE_FIXED32", "TYPE_FIXED64",
		"TYPE_SFIXED32", "TYPE_SFIXED64":
		return "integer"
	case "TYPE_BOOL":
		return "boolean"
	case "TYPE_STRING":
		return "string"
	case "TYPE_BYTES":
		return "string" // base64 encoded
	case "TYPE_MESSAGE":
		return "object"
	case "TYPE_ENUM":
		return "string"
	default:
		return "string"
	}
}

// extractComments extracts leading comments from source code info
func extractComments(info *descriptorpb.SourceCodeInfo_Location) string {
	if info == nil {
		return ""
	}
	if info.LeadingComments != nil {
		return *info.LeadingComments
	}
	return ""
}

// Clear removes all registered descriptors
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.files = make(map[string]*desc.FileDescriptor)
	r.services = make(map[string]*desc.ServiceDescriptor)
	r.messages = make(map[string]*desc.MessageDescriptor)
}

// Stats returns statistics about the registry
type Stats struct {
	FileCount    int
	ServiceCount int
	MessageCount int
}

// GetStats returns current registry statistics
func (r *Registry) GetStats() Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return Stats{
		FileCount:    len(r.files),
		ServiceCount: len(r.services),
		MessageCount: len(r.messages),
	}
}

// HasService checks if a service is registered
func (r *Registry) HasService(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.services[name]
	return exists
}

// ParseError wraps descriptor parsing errors
type ParseError struct {
	File    string
	Message string
	Cause   error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error in %s: %s", e.File, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

// ValidateDescriptors performs basic validation on file descriptors
func ValidateDescriptors(fds *descriptorpb.FileDescriptorSet) error {
	if fds == nil {
		return fmt.Errorf("nil file descriptor set")
	}

	if len(fds.File) == 0 {
		return fmt.Errorf("empty file descriptor set")
	}

	// Check for basic consistency
	fileNames := make(map[string]bool)
	for _, file := range fds.File {
		name := file.GetName()
		if name == "" {
			return fmt.Errorf("file descriptor with empty name")
		}

		if fileNames[name] {
			return fmt.Errorf("duplicate file descriptor: %s", name)
		}
		fileNames[name] = true
	}

	return nil
}

// NewFromParser creates a registry from parsed proto files (alternative construction)
func NewFromParser(parser *protoparse.Parser, filenames ...string) (*Registry, error) {
	fds, err := parser.ParseFiles(filenames...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse files: %w", err)
	}

	registry := New()

	// Convert jhump descriptors to descriptorpb format
	fdSet := &descriptorpb.FileDescriptorSet{
		File: make([]*descriptorpb.FileDescriptorProto, len(fds)),
	}

	for i, fd := range fds {
		fdSet.File[i] = fd.AsFileDescriptorProto()
	}

	if err := registry.Register(fdSet); err != nil {
		return nil, fmt.Errorf("failed to register descriptors: %w", err)
	}

	return registry, nil
}

// Clone creates a deep copy of the registry
func (r *Registry) Clone() *Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := New()
	clone.files = make(map[string]*desc.FileDescriptor, len(r.files))
	clone.services = make(map[string]*desc.ServiceDescriptor, len(r.services))
	clone.messages = make(map[string]*desc.MessageDescriptor, len(r.messages))

	for k, v := range r.files {
		clone.files[k] = v
	}
	for k, v := range r.services {
		clone.services[k] = v
	}
	for k, v := range r.messages {
		clone.messages[k] = v
	}

	return clone
}

// MarshalBinary serializes the registry to binary format
func (r *Registry) MarshalBinary() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fds := &descriptorpb.FileDescriptorSet{
		File: make([]*descriptorpb.FileDescriptorProto, 0, len(r.files)),
	}

	for _, fd := range r.files {
		fds.File = append(fds.File, fd.AsFileDescriptorProto())
	}

	return proto.Marshal(fds)
}

// UnmarshalBinary deserializes a registry from binary format
func (r *Registry) UnmarshalBinary(data []byte) error {
	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(data, fds); err != nil {
		return fmt.Errorf("failed to unmarshal descriptor set: %w", err)
	}

	return r.Register(fds)
}
