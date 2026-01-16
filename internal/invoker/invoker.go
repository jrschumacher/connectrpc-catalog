package invoker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	catalogv1 "github.com/opentdf/connectrpc-catalog/gen/catalog/v1"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Invoker handles dynamic gRPC invocations using descriptor-based reflection
type Invoker struct {
	// Connection pool for reusing gRPC connections
	connections map[string]*grpc.ClientConn
	// HTTP client for Connect protocol
	httpClient *http.Client
}

// New creates a new Invoker instance
func New() *Invoker {
	return &Invoker{
		connections: make(map[string]*grpc.ClientConn),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// InvokeRequest contains parameters for a dynamic gRPC invocation
type InvokeRequest struct {
	Endpoint        string
	ServiceName     string
	MethodName      string
	RequestJSON     json.RawMessage
	UseTLS          bool
	ServerName      string
	TimeoutSeconds  int32
	Metadata        map[string]string
	MethodDesc      *desc.MethodDescriptor
	Transport       catalogv1.Transport // Transport protocol to use
}

// InvokeResponse contains the result of a gRPC invocation
type InvokeResponse struct {
	Success       bool
	ResponseJSON  json.RawMessage
	Error         string
	Metadata      map[string]string
	StatusCode    int32
	StatusMessage string
}

// InvokeUnary performs a unary call using the specified transport
func (inv *Invoker) InvokeUnary(ctx context.Context, req InvokeRequest) (*InvokeResponse, error) {
	// Route based on transport (default to Connect when unspecified/zero value)
	switch req.Transport {
	case catalogv1.Transport_TRANSPORT_GRPC:
		return inv.invokeGRPC(ctx, req)
	case catalogv1.Transport_TRANSPORT_GRPC_WEB:
		// gRPC-Web not yet supported, fall back to Connect
		return inv.invokeConnect(ctx, req)
	default:
		// TRANSPORT_CONNECT (0) or any unspecified value defaults to Connect
		return inv.invokeConnect(ctx, req)
	}
}

// invokeConnect performs a unary call using the Connect protocol (HTTP/JSON)
func (inv *Invoker) invokeConnect(ctx context.Context, req InvokeRequest) (*InvokeResponse, error) {
	// Build the Connect URL: http(s)://{endpoint}/{service}/{method}
	scheme := "http"
	if req.UseTLS {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, req.Endpoint, req.ServiceName, req.MethodName)

	// Create HTTP request with the JSON body
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(req.RequestJSON))
	if err != nil {
		return &InvokeResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	// Set Connect protocol headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Connect-Protocol-Version", "1")

	// Add custom metadata headers
	for k, v := range req.Metadata {
		httpReq.Header.Set(k, v)
	}

	// Create a client with timeout
	client := inv.httpClient
	if req.TimeoutSeconds > 0 {
		client = &http.Client{
			Timeout: time.Duration(req.TimeoutSeconds) * time.Second,
		}
		if req.UseTLS {
			client.Transport = &http.Transport{
				TLSClientConfig: &tls.Config{
					ServerName: req.ServerName,
				},
			}
		}
	}

	// Execute the request
	resp, err := client.Do(httpReq)
	if err != nil {
		return &InvokeResponse{
			Success: false,
			Error:   fmt.Sprintf("request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &InvokeResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to read response: %v", err),
		}, nil
	}

	// Collect response headers as metadata
	respMetadata := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respMetadata[k] = v[0]
		}
	}

	// Check for Connect error response
	if resp.StatusCode != http.StatusOK {
		// Try to parse Connect error format
		var connectErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &connectErr) == nil && connectErr.Message != "" {
			return &InvokeResponse{
				Success:       false,
				Error:         connectErr.Message,
				StatusCode:    int32(resp.StatusCode),
				StatusMessage: connectErr.Code,
				Metadata:      respMetadata,
			}, nil
		}
		return &InvokeResponse{
			Success:       false,
			Error:         fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			StatusCode:    int32(resp.StatusCode),
			StatusMessage: resp.Status,
			Metadata:      respMetadata,
		}, nil
	}

	return &InvokeResponse{
		Success:       true,
		ResponseJSON:  body,
		StatusCode:    0,
		StatusMessage: "OK",
		Metadata:      respMetadata,
	}, nil
}

// invokeGRPC performs a unary gRPC call using dynamic invocation
func (inv *Invoker) invokeGRPC(ctx context.Context, req InvokeRequest) (*InvokeResponse, error) {
	// Validate method descriptor
	if req.MethodDesc == nil {
		return nil, fmt.Errorf("method descriptor is required for gRPC transport")
	}

	if req.MethodDesc.IsClientStreaming() || req.MethodDesc.IsServerStreaming() {
		return nil, fmt.Errorf("streaming methods not supported (use InvokeUnary for unary RPCs only)")
	}

	// Get or create gRPC connection
	conn, err := inv.getConnection(req.Endpoint, req.UseTLS, req.ServerName)
	if err != nil {
		return &InvokeResponse{
			Success: false,
			Error:   fmt.Sprintf("connection failed: %v", err),
		}, nil
	}

	// Create dynamic stub
	stub := grpcdynamic.NewStub(conn)

	// Parse request JSON into dynamic message
	reqMsg := dynamic.NewMessage(req.MethodDesc.GetInputType())

	unmarshaler := &jsonpb.Unmarshaler{}
	if err := reqMsg.UnmarshalJSONPB(unmarshaler, []byte(req.RequestJSON)); err != nil {
		return &InvokeResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request JSON: %v", err),
		}, nil
	}

	// Setup context with timeout and metadata
	invokeCtx := ctx
	if req.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		invokeCtx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	// Add request metadata
	if len(req.Metadata) > 0 {
		md := metadata.New(req.Metadata)
		invokeCtx = metadata.NewOutgoingContext(invokeCtx, md)
	}

	// Prepare response metadata capture
	var respHeader, respTrailer metadata.MD

	// Invoke the method
	respMsg, err := stub.InvokeRpc(invokeCtx, req.MethodDesc, reqMsg,
		grpc.Header(&respHeader),
		grpc.Trailer(&respTrailer),
	)

	// Handle invocation error
	if err != nil {
		statusCode, statusMsg := extractGRPCStatus(err)
		return &InvokeResponse{
			Success:       false,
			Error:         err.Error(),
			StatusCode:    statusCode,
			StatusMessage: statusMsg,
			Metadata:      mergeMetadata(respHeader, respTrailer),
		}, nil
	}

	// Convert response to JSON - respMsg is already a *dynamic.Message
	dynRespMsg, ok := respMsg.(*dynamic.Message)
	if !ok {
		return &InvokeResponse{
			Success: false,
			Error:   "response is not a dynamic message",
		}, nil
	}

	marshaler := &jsonpb.Marshaler{}
	respJSON, err := dynRespMsg.MarshalJSONPB(marshaler)
	if err != nil {
		return &InvokeResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to marshal response: %v", err),
		}, nil
	}

	return &InvokeResponse{
		Success:       true,
		ResponseJSON:  respJSON,
		StatusCode:    0, // OK
		StatusMessage: "OK",
		Metadata:      mergeMetadata(respHeader, respTrailer),
	}, nil
}

// getConnection retrieves or creates a gRPC connection
func (inv *Invoker) getConnection(endpoint string, useTLS bool, serverName string) (*grpc.ClientConn, error) {
	// Check if connection already exists
	connKey := fmt.Sprintf("%s:%v:%s", endpoint, useTLS, serverName)
	if conn, exists := inv.connections[connKey]; exists {
		// Verify connection is still valid
		if conn.GetState().String() != "SHUTDOWN" {
			return conn, nil
		}
		// Connection is dead, remove it
		delete(inv.connections, connKey)
	}

	// Create new connection
	var opts []grpc.DialOption

	if useTLS {
		// Use TLS credentials
		tlsConfig := &tls.Config{}
		if serverName != "" {
			// Override server name for TLS verification
			tlsConfig.ServerName = serverName
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		// Use insecure credentials
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Use blocking dial with short timeout for fast failure when server is unreachable
	// This ensures behavior similar to HTTP connect failures
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()

	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.DialContext(dialCtx, endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", endpoint, err)
	}

	// Cache the connection
	inv.connections[connKey] = conn

	return conn, nil
}

// Close closes all open gRPC connections
func (inv *Invoker) Close() error {
	var errs []error
	for key, conn := range inv.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection %s: %w", key, err))
		}
	}

	inv.connections = make(map[string]*grpc.ClientConn)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// extractGRPCStatus extracts status code and message from gRPC error
func extractGRPCStatus(err error) (int32, string) {
	if err == nil {
		return 0, "OK"
	}

	// Try to extract gRPC status
	// This is simplified - in production use google.golang.org/grpc/status
	// For now, return generic error code
	return 2, err.Error() // 2 = UNKNOWN
}

// mergeMetadata combines header and trailer metadata
func mergeMetadata(header, trailer metadata.MD) map[string]string {
	result := make(map[string]string)

	for k, v := range header {
		if len(v) > 0 {
			result[k] = v[0] // Take first value
		}
	}

	for k, v := range trailer {
		if len(v) > 0 {
			result["trailer-"+k] = v[0] // Prefix trailer keys
		}
	}

	return result
}

// InvokeUnarySimple is a simplified version that takes raw parameters
// This is a convenience wrapper around InvokeUnary
func InvokeUnarySimple(
	ctx context.Context,
	endpoint string,
	serviceName string,
	methodName string,
	methodDesc *desc.MethodDescriptor,
	reqJSON json.RawMessage,
) (json.RawMessage, error) {
	inv := New()
	defer inv.Close()

	req := InvokeRequest{
		Endpoint:       endpoint,
		ServiceName:    serviceName,
		MethodName:     methodName,
		RequestJSON:    reqJSON,
		UseTLS:         false,
		TimeoutSeconds: 30,
		MethodDesc:     methodDesc,
	}

	resp, err := inv.InvokeUnary(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("invocation failed: %s", resp.Error)
	}

	return resp.ResponseJSON, nil
}

// ValidateRequest checks if an invocation request is valid
func ValidateRequest(req InvokeRequest) error {
	if req.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	if req.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if req.MethodName == "" {
		return fmt.Errorf("method name is required")
	}

	if req.MethodDesc == nil {
		return fmt.Errorf("method descriptor is required")
	}

	if len(req.RequestJSON) == 0 {
		return fmt.Errorf("request JSON is required")
	}

	// Validate JSON is well-formed
	var tmp interface{}
	if err := json.Unmarshal(req.RequestJSON, &tmp); err != nil {
		return fmt.Errorf("invalid request JSON: %w", err)
	}

	return nil
}

// ConnectionStats provides statistics about active connections
type ConnectionStats struct {
	TotalConnections int
	ActiveConnections int
	EndpointCounts   map[string]int
}

// GetConnectionStats returns statistics about the invoker's connections
func (inv *Invoker) GetConnectionStats() ConnectionStats {
	stats := ConnectionStats{
		TotalConnections:  len(inv.connections),
		ActiveConnections: 0,
		EndpointCounts:    make(map[string]int),
	}

	for key, conn := range inv.connections {
		state := conn.GetState()
		if state.String() != "SHUTDOWN" && state.String() != "TRANSIENT_FAILURE" {
			stats.ActiveConnections++
		}
		stats.EndpointCounts[key]++
	}

	return stats
}

// CloseConnection closes a specific connection by endpoint
func (inv *Invoker) CloseConnection(endpoint string, useTLS bool, serverName string) error {
	connKey := fmt.Sprintf("%s:%v:%s", endpoint, useTLS, serverName)

	conn, exists := inv.connections[connKey]
	if !exists {
		return fmt.Errorf("connection not found: %s", connKey)
	}

	if err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	delete(inv.connections, connKey)
	return nil
}

// WaitForReady waits for a connection to be ready
func (inv *Invoker) WaitForReady(ctx context.Context, endpoint string, useTLS bool, serverName string) error {
	conn, err := inv.getConnection(endpoint, useTLS, serverName)
	if err != nil {
		return err
	}

	// Wait for connection to be ready
	for {
		state := conn.GetState()
		if state.String() == "READY" {
			return nil
		}
		if state.String() == "SHUTDOWN" || state.String() == "TRANSIENT_FAILURE" {
			return fmt.Errorf("connection failed: state=%s", state.String())
		}

		// Wait a bit before checking again
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Continue waiting
		}
	}
}
