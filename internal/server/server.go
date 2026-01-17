package server

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"
	"github.com/opentdf/connectrpc-catalog/internal/invoker"
	"github.com/opentdf/connectrpc-catalog/internal/loader"
	"github.com/opentdf/connectrpc-catalog/internal/session"
	"google.golang.org/protobuf/types/descriptorpb"
)

// CatalogServer implements the CatalogService ConnectRPC handlers
type CatalogServer struct {
	sessionManager *session.Manager
}

// New creates a new CatalogServer instance
func New() *CatalogServer {
	return &CatalogServer{
		sessionManager: session.NewManager(session.DefaultSessionTTL),
	}
}

// LoadProtos implements the LoadProtos RPC handler
func (s *CatalogServer) LoadProtos(
	ctx context.Context,
	req *connect.Request[catalogv1.LoadProtosRequest],
) (*connect.Response[catalogv1.LoadProtosResponse], error) {
	// Get or create session
	sessionID := req.Header().Get("X-Session-ID")
	state, newSessionID, err := s.sessionManager.GetOrCreate(sessionID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Determine the source type and load descriptors
	var fds *descriptorpb.FileDescriptorSet

	switch source := req.Msg.Source.(type) {
	case *catalogv1.LoadProtosRequest_ProtoPath:
		fds, err = loader.LoadFromPath(source.ProtoPath)
		if err != nil {
			resp := connect.NewResponse(&catalogv1.LoadProtosResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to load from path: %v", err),
			})
			resp.Header().Set("X-Session-ID", newSessionID)
			return resp, nil
		}

	case *catalogv1.LoadProtosRequest_ProtoRepo:
		fds, err = loader.LoadFromGitHub(source.ProtoRepo)
		if err != nil {
			resp := connect.NewResponse(&catalogv1.LoadProtosResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to load from GitHub: %v", err),
			})
			resp.Header().Set("X-Session-ID", newSessionID)
			return resp, nil
		}

	case *catalogv1.LoadProtosRequest_BufModule:
		fds, err = loader.LoadFromBufModule(source.BufModule)
		if err != nil {
			resp := connect.NewResponse(&catalogv1.LoadProtosResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to load from Buf module: %v", err),
			})
			resp.Header().Set("X-Session-ID", newSessionID)
			return resp, nil
		}

	case *catalogv1.LoadProtosRequest_ReflectionEndpoint:
		// Build reflection options from request
		opts := loader.ReflectionOptions{
			UseTLS:         true, // Default to TLS
			TimeoutSeconds: 10,   // Default timeout
		}
		if refOpts := req.Msg.GetReflectionOptions(); refOpts != nil {
			opts.UseTLS = refOpts.GetUseTls()
			opts.ServerName = refOpts.GetServerName()
			if refOpts.GetTimeoutSeconds() > 0 {
				opts.TimeoutSeconds = refOpts.GetTimeoutSeconds()
			}
		}

		fds, err = loader.LoadFromReflection(source.ReflectionEndpoint, opts)
		if err != nil {
			resp := connect.NewResponse(&catalogv1.LoadProtosResponse{
				Success: false,
				Error:   fmt.Sprintf("failed to load from reflection: %v", err),
			})
			resp.Header().Set("X-Session-ID", newSessionID)
			return resp, nil
		}

	default:
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("no source specified in request"),
		)
	}

	// Register the loaded descriptors using session registry
	if err := state.Registry.Register(fds); err != nil {
		resp := connect.NewResponse(&catalogv1.LoadProtosResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to register descriptors: %v", err),
		})
		resp.Header().Set("X-Session-ID", newSessionID)
		return resp, nil
	}

	// Get statistics
	info := loader.GetDescriptorInfo(fds)

	resp := connect.NewResponse(&catalogv1.LoadProtosResponse{
		Success:      true,
		ServiceCount: int32(len(info.Services)),
		FileCount:    int32(info.Files),
	})
	resp.Header().Set("X-Session-ID", newSessionID)
	return resp, nil
}

// ListServices implements the ListServices RPC handler
func (s *CatalogServer) ListServices(
	ctx context.Context,
	req *connect.Request[catalogv1.ListServicesRequest],
) (*connect.Response[catalogv1.ListServicesResponse], error) {
	// Get or create session
	sessionID := req.Header().Get("X-Session-ID")
	state, newSessionID, err := s.sessionManager.GetOrCreate(sessionID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Get all services from session registry
	services := state.Registry.ListServices()

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

	resp := connect.NewResponse(&catalogv1.ListServicesResponse{
		Services: protoServices,
	})
	resp.Header().Set("X-Session-ID", newSessionID)
	return resp, nil
}

// GetServiceSchema implements the GetServiceSchema RPC handler
func (s *CatalogServer) GetServiceSchema(
	ctx context.Context,
	req *connect.Request[catalogv1.GetServiceSchemaRequest],
) (*connect.Response[catalogv1.GetServiceSchemaResponse], error) {
	// Get or create session
	sessionID := req.Header().Get("X-Session-ID")
	state, newSessionID, err := s.sessionManager.GetOrCreate(sessionID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	serviceName := req.Msg.ServiceName

	if serviceName == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("service_name is required"),
		)
	}

	// Get service schema from session registry
	serviceInfo, messageSchemas, err := state.Registry.GetServiceSchema(serviceName)
	if err != nil {
		resp := connect.NewResponse(&catalogv1.GetServiceSchemaResponse{
			Error: fmt.Sprintf("failed to get service schema: %v", err),
		})
		resp.Header().Set("X-Session-ID", newSessionID)
		return resp, nil
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

	resp := connect.NewResponse(&catalogv1.GetServiceSchemaResponse{
		Service:        protoServiceInfo,
		MessageSchemas: messageSchemas,
	})
	resp.Header().Set("X-Session-ID", newSessionID)
	return resp, nil
}

// InvokeGRPC implements the InvokeGRPC RPC handler
func (s *CatalogServer) InvokeGRPC(
	ctx context.Context,
	req *connect.Request[catalogv1.InvokeGRPCRequest],
) (*connect.Response[catalogv1.InvokeGRPCResponse], error) {
	// Get or create session
	sessionID := req.Header().Get("X-Session-ID")
	state, newSessionID, err := s.sessionManager.GetOrCreate(sessionID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

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

	// Get method descriptor from session registry
	methodDesc, err := state.Registry.GetMethodDescriptor(req.Msg.Service, req.Msg.Method)
	if err != nil {
		resp := connect.NewResponse(&catalogv1.InvokeGRPCResponse{
			Success: false,
			Error:   fmt.Sprintf("method not found: %v", err),
		})
		resp.Header().Set("X-Session-ID", newSessionID)
		return resp, nil
	}

	// Check for streaming methods (not supported in MVP)
	if methodDesc.IsClientStreaming() || methodDesc.IsServerStreaming() {
		resp := connect.NewResponse(&catalogv1.InvokeGRPCResponse{
			Success: false,
			Error:   "streaming methods are not supported in MVP (unary only)",
		})
		resp.Header().Set("X-Session-ID", newSessionID)
		return resp, nil
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
		Transport:      req.Msg.Transport,
	}

	// Perform invocation using session invoker
	invokeResp, err := state.Invoker.InvokeUnary(ctx, invokeReq)
	if err != nil {
		resp := connect.NewResponse(&catalogv1.InvokeGRPCResponse{
			Success: false,
			Error:   fmt.Sprintf("invocation error: %v", err),
		})
		resp.Header().Set("X-Session-ID", newSessionID)
		return resp, nil
	}

	// Convert response
	resp := connect.NewResponse(&catalogv1.InvokeGRPCResponse{
		Success:       invokeResp.Success,
		ResponseJson:  string(invokeResp.ResponseJSON),
		Error:         invokeResp.Error,
		Metadata:      invokeResp.Metadata,
		StatusCode:    invokeResp.StatusCode,
		StatusMessage: invokeResp.StatusMessage,
	})
	resp.Header().Set("X-Session-ID", newSessionID)
	return resp, nil
}

// Close releases all resources held by the server
func (s *CatalogServer) Close() error {
	if s.sessionManager != nil {
		s.sessionManager.Close()
	}
	return nil
}

// GetSessionManager returns the session manager (for testing/inspection)
func (s *CatalogServer) GetSessionManager() *session.Manager {
	return s.sessionManager
}

// Stats returns server statistics
type Stats struct {
	SessionStats session.Stats
}

// GetStats returns current server statistics
func (s *CatalogServer) GetStats() Stats {
	return Stats{
		SessionStats: s.sessionManager.GetStats(),
	}
}

// ValidateSetup checks if the server is properly configured
func (s *CatalogServer) ValidateSetup() error {
	if s.sessionManager == nil {
		return fmt.Errorf("session manager is nil")
	}
	return nil
}
