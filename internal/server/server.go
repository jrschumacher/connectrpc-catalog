package server

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
	"github.com/opentdf/connectrpc-catalog/internal/invoker"
	"github.com/opentdf/connectrpc-catalog/internal/loader"
	"github.com/opentdf/connectrpc-catalog/internal/registry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// CatalogServer implements the CatalogService ConnectRPC handlers
type CatalogServer struct {
	registry *registry.Registry
	invoker  *invoker.Invoker
}

// New creates a new CatalogServer instance
func New() *CatalogServer {
	return &CatalogServer{
		registry: registry.New(),
		invoker:  invoker.New(),
	}
}

// LoadProtos implements the LoadProtos RPC handler
func (s *CatalogServer) LoadProtos(
	ctx context.Context,
	req *connect.Request[catalogv1.LoadProtosRequest],
) (*connect.Response[catalogv1.LoadProtosResponse], error) {
	// Determine the source type and load descriptors
	var fds *descriptorpb.FileDescriptorSet
	var err error

	switch source := req.Msg.Source.(type) {
	case *catalogv1.LoadProtosRequest_ProtoPath:
		fds, err = loader.LoadFromPath(source.ProtoPath)
		if err != nil {
			return connect.NewResponse(&catalogv1.LoadProtosResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to load from path: %v", err),
			}), nil
		}

	case *catalogv1.LoadProtosRequest_ProtoRepo:
		fds, err = loader.LoadFromGitHub(source.ProtoRepo)
		if err != nil {
			return connect.NewResponse(&catalogv1.LoadProtosResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to load from GitHub: %v", err),
			}), nil
		}

	case *catalogv1.LoadProtosRequest_BufModule:
		fds, err = loader.LoadFromBufModule(source.BufModule)
		if err != nil {
			return connect.NewResponse(&catalogv1.LoadProtosResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to load from Buf module: %v", err),
			}), nil
		}

	default:
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("no source specified in request"),
		)
	}

	// Register the loaded descriptors
	if err := s.registry.Register(fds); err != nil {
		return connect.NewResponse(&catalogv1.LoadProtosResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to register descriptors: %v", err),
		}), nil
	}

	// Get statistics
	info := loader.GetDescriptorInfo(fds)

	return connect.NewResponse(&catalogv1.LoadProtosResponse{
		Success:      true,
		ServiceCount: int32(len(info.Services)),
		FileCount:    int32(info.Files),
	}), nil
}

// ListServices implements the ListServices RPC handler
func (s *CatalogServer) ListServices(
	ctx context.Context,
	req *connect.Request[catalogv1.ListServicesRequest],
) (*connect.Response[catalogv1.ListServicesResponse], error) {
	// Get all services from registry
	services := s.registry.ListServices()

	// Convert to proto response format
	protoServices := make([]*catalogv1.ServiceInfo, len(services))
	for i, svc := range services {
		methods := make([]*catalogv1.MethodInfo, len(svc.Methods))
		for j, method := range svc.Methods {
			methods[j] = &catalogv1.MethodInfo{
				Name:            method.Name,
				InputType:       method.InputType,
				OutputType:      method.OutputType,
				Documentation:   method.Documentation,
				ClientStreaming: method.ClientStreaming,
				ServerStreaming: method.ServerStreaming,
			}
		}

		protoServices[i] = &catalogv1.ServiceInfo{
			Name:          svc.Name,
			Package:       svc.Package,
			Methods:       methods,
			Documentation: svc.Documentation,
		}
	}

	return connect.NewResponse(&catalogv1.ListServicesResponse{
		Services: protoServices,
	}), nil
}

// GetServiceSchema implements the GetServiceSchema RPC handler
func (s *CatalogServer) GetServiceSchema(
	ctx context.Context,
	req *connect.Request[catalogv1.GetServiceSchemaRequest],
) (*connect.Response[catalogv1.GetServiceSchemaResponse], error) {
	serviceName := req.Msg.ServiceName

	if serviceName == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("service_name is required"),
		)
	}

	// Get service schema from registry
	serviceInfo, messageSchemas, err := s.registry.GetServiceSchema(serviceName)
	if err != nil {
		return connect.NewResponse(&catalogv1.GetServiceSchemaResponse{
			Error: fmt.Sprintf("failed to get service schema: %v", err),
		}), nil
	}

	// Convert service info to proto format
	methods := make([]*catalogv1.MethodInfo, len(serviceInfo.Methods))
	for i, method := range serviceInfo.Methods {
		methods[i] = &catalogv1.MethodInfo{
			Name:            method.Name,
			InputType:       method.InputType,
			OutputType:      method.OutputType,
			Documentation:   method.Documentation,
			ClientStreaming: method.ClientStreaming,
			ServerStreaming: method.ServerStreaming,
		}
	}

	protoServiceInfo := &catalogv1.ServiceInfo{
		Name:          serviceInfo.Name,
		Package:       serviceInfo.Package,
		Methods:       methods,
		Documentation: serviceInfo.Documentation,
	}

	return connect.NewResponse(&catalogv1.GetServiceSchemaResponse{
		Service:        protoServiceInfo,
		MessageSchemas: messageSchemas,
	}), nil
}

// InvokeGRPC implements the InvokeGRPC RPC handler
func (s *CatalogServer) InvokeGRPC(
	ctx context.Context,
	req *connect.Request[catalogv1.InvokeGRPCRequest],
) (*connect.Response[catalogv1.InvokeGRPCResponse], error) {
	// Validate required fields
	if req.Msg.Endpoint == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("endpoint is required"),
		)
	}
	if req.Msg.Service == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("service is required"),
		)
	}
	if req.Msg.Method == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("method is required"),
		)
	}

	// Get method descriptor from registry
	methodDesc, err := s.registry.GetMethodDescriptor(req.Msg.Service, req.Msg.Method)
	if err != nil {
		return connect.NewResponse(&catalogv1.InvokeGRPCResponse{
			Success: false,
			Error:   fmt.Sprintf("method not found: %v", err),
		}), nil
	}

	// Check for streaming methods (not supported in MVP)
	if methodDesc.IsClientStreaming() || methodDesc.IsServerStreaming() {
		return connect.NewResponse(&catalogv1.InvokeGRPCResponse{
			Success: false,
			Error:   "streaming methods are not supported in MVP (unary only)",
		}), nil
	}

	// Parse request JSON
	var requestJSON json.RawMessage
	if req.Msg.RequestJson != "" {
		requestJSON = json.RawMessage(req.Msg.RequestJson)
	} else {
		requestJSON = json.RawMessage("{}")
	}

	// Set default timeout if not specified
	timeoutSeconds := req.Msg.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	// Build invocation request
	invokeReq := invoker.InvokeRequest{
		Endpoint:       req.Msg.Endpoint,
		ServiceName:    req.Msg.Service,
		MethodName:     req.Msg.Method,
		RequestJSON:    requestJSON,
		UseTLS:         req.Msg.UseTls,
		ServerName:     req.Msg.ServerName,
		TimeoutSeconds: timeoutSeconds,
		Metadata:       req.Msg.Metadata,
		MethodDesc:     methodDesc,
	}

	// Perform invocation
	invokeResp, err := s.invoker.InvokeUnary(ctx, invokeReq)
	if err != nil {
		return connect.NewResponse(&catalogv1.InvokeGRPCResponse{
			Success: false,
			Error:   fmt.Sprintf("invocation error: %v", err),
		}), nil
	}

	// Convert response
	return connect.NewResponse(&catalogv1.InvokeGRPCResponse{
		Success:       invokeResp.Success,
		ResponseJson:  string(invokeResp.ResponseJSON),
		Error:         invokeResp.Error,
		Metadata:      invokeResp.Metadata,
		StatusCode:    invokeResp.StatusCode,
		StatusMessage: invokeResp.StatusMessage,
	}), nil
}

// Close releases all resources held by the server
func (s *CatalogServer) Close() error {
	if s.invoker != nil {
		return s.invoker.Close()
	}
	return nil
}

// GetRegistry returns the underlying registry (for testing/inspection)
func (s *CatalogServer) GetRegistry() *registry.Registry {
	return s.registry
}

// GetInvoker returns the underlying invoker (for testing/inspection)
func (s *CatalogServer) GetInvoker() *invoker.Invoker {
	return s.invoker
}

// Stats returns server statistics
type Stats struct {
	RegistryStats       registry.Stats
	ConnectionStats     invoker.ConnectionStats
	TotalLoadedServices int
	TotalInvocations    int
}

// GetStats returns current server statistics
func (s *CatalogServer) GetStats() Stats {
	return Stats{
		RegistryStats:   s.registry.GetStats(),
		ConnectionStats: s.invoker.GetConnectionStats(),
	}
}

// ClearRegistry clears all loaded descriptors from the registry
func (s *CatalogServer) ClearRegistry() {
	s.registry.Clear()
}

// ValidateSetup checks if the server is properly configured
func (s *CatalogServer) ValidateSetup() error {
	if s.registry == nil {
		return fmt.Errorf("registry is nil")
	}
	if s.invoker == nil {
		return fmt.Errorf("invoker is nil")
	}
	return nil
}
