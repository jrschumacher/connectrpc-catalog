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

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// DefaultMaxConnections is the default maximum number of cached connections
	DefaultMaxConnections = 100
	// DefaultConnectionTTL is the default time-to-live for cached connections
	DefaultConnectionTTL = 5 * time.Minute
	// ConnectionIdleTimeout is the timeout for idle connections
	ConnectionIdleTimeout = 2 * time.Minute
)

// connectionMetadata tracks metadata about a cached connection
type connectionMetadata struct {
	conn      *grpc.ClientConn
	createdAt time.Time
	lastUsed  time.Time
}

// Invoker handles dynamic gRPC invocations using descriptor-based reflection
type Invoker struct {
	// Connection pool for reusing gRPC connections with metadata
	connections map[string]*connectionMetadata
	// HTTP client for Connect protocol
	httpClient *http.Client
	// Maximum number of connections to cache
	maxConnections int
	// Connection time-to-live
	connectionTTL time.Duration
}

// New creates a new Invoker instance with default connection pool settings
func New() *Invoker {
	return &Invoker{
		connections:    make(map[string]*connectionMetadata),
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		maxConnections: DefaultMaxConnections,
		connectionTTL:  DefaultConnectionTTL,
	}
}

// NewWithLimits creates a new Invoker with custom connection pool limits
func NewWithLimits(maxConnections int, ttl time.Duration) *Invoker {
	return &Invoker{
		connections:    make(map[string]*connectionMetadata),
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		maxConnections: maxConnections,
		connectionTTL:  ttl,
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

	if err := reqMsg.UnmarshalJSON(req.RequestJSON); err != nil {
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

	respJSON, err := dynRespMsg.MarshalJSON()
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

// getConnection retrieves or creates a gRPC connection with pool management
func (inv *Invoker) getConnection(endpoint string, useTLS bool, serverName string) (*grpc.ClientConn, error) {
	connKey := fmt.Sprintf("%s:%v:%s", endpoint, useTLS, serverName)
	now := time.Now()

	// Clean up stale connections before checking pool
	inv.cleanupStaleConnections()

	// Check if connection already exists and is valid
	if connMeta, exists := inv.connections[connKey]; exists {
		// Check if connection is still valid and not expired
		if connMeta.conn.GetState().String() != "SHUTDOWN" &&
			now.Sub(connMeta.createdAt) < inv.connectionTTL {
			// Update last used time
			connMeta.lastUsed = now
			return connMeta.conn, nil
		}
		// Connection is dead or expired, remove it
		_ = connMeta.conn.Close()
		delete(inv.connections, connKey)
	}

	// Enforce maximum connection limit
	if len(inv.connections) >= inv.maxConnections {
		inv.evictOldestConnection()
	}

	// Create new connection
	var opts []grpc.DialOption

	if useTLS {
		tlsConfig := &tls.Config{}
		if serverName != "" {
			tlsConfig.ServerName = serverName
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Use blocking dial with short timeout for fast failure when server is unreachable
	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer dialCancel()

	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.DialContext(dialCtx, endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", endpoint, err)
	}

	// Cache the connection with metadata
	inv.connections[connKey] = &connectionMetadata{
		conn:      conn,
		createdAt: now,
		lastUsed:  now,
	}

	return conn, nil
}

// cleanupStaleConnections removes expired or idle connections from the pool
func (inv *Invoker) cleanupStaleConnections() {
	now := time.Now()
	for key, connMeta := range inv.connections {
		// Check if connection has expired or been idle too long
		if now.Sub(connMeta.createdAt) >= inv.connectionTTL ||
			now.Sub(connMeta.lastUsed) >= ConnectionIdleTimeout ||
			connMeta.conn.GetState().String() == "SHUTDOWN" {
			_ = connMeta.conn.Close()
			delete(inv.connections, key)
		}
	}
}

// evictOldestConnection removes the least recently used connection
func (inv *Invoker) evictOldestConnection() {
	var oldestKey string
	var oldestTime time.Time

	for key, connMeta := range inv.connections {
		if oldestKey == "" || connMeta.lastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = connMeta.lastUsed
		}
	}

	if oldestKey != "" {
		if connMeta, exists := inv.connections[oldestKey]; exists {
			_ = connMeta.conn.Close()
			delete(inv.connections, oldestKey)
		}
	}
}

// Close closes all open gRPC connections
func (inv *Invoker) Close() error {
	var errs []error
	for key, connMeta := range inv.connections {
		if err := connMeta.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection %s: %w", key, err))
		}
	}

	inv.connections = make(map[string]*connectionMetadata)

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

	// Extract gRPC status using the proper status package
	if st, ok := status.FromError(err); ok {
		return int32(st.Code()), st.Message()
	}

	// Fallback to generic error if status extraction fails
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

	for key, connMeta := range inv.connections {
		state := connMeta.conn.GetState()
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

	connMeta, exists := inv.connections[connKey]
	if !exists {
		return fmt.Errorf("connection not found: %s", connKey)
	}

	if err := connMeta.conn.Close(); err != nil {
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
