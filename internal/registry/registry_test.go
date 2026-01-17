package registry

import (
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"
)

// createTestFileDescriptorSet creates a minimal FileDescriptorSet for testing
func createTestFileDescriptorSet() *descriptorpb.FileDescriptorSet {
	serviceName := "TestService"
	methodName := "TestMethod"
	packageName := "test.v1"

	inputType := ".test.v1.TestRequest"
	outputType := ".test.v1.TestResponse"

	// Create method descriptor
	method := &descriptorpb.MethodDescriptorProto{
		Name:       &methodName,
		InputType:  &inputType,
		OutputType: &outputType,
	}

	// Create service descriptor
	service := &descriptorpb.ServiceDescriptorProto{
		Name:   &serviceName,
		Method: []*descriptorpb.MethodDescriptorProto{method},
	}

	// Create message descriptors
	requestMsgName := "TestRequest"
	responseMsgName := "TestResponse"

	requestField1Name := "name"
	requestField1Number := int32(1)
	requestField1Type := descriptorpb.FieldDescriptorProto_TYPE_STRING
	requestField1Label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	requestMsg := &descriptorpb.DescriptorProto{
		Name: &requestMsgName,
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   &requestField1Name,
				Number: &requestField1Number,
				Type:   &requestField1Type,
				Label:  &requestField1Label,
			},
		},
	}

	responseField1Name := "message"
	responseField1Number := int32(1)
	responseField1Type := descriptorpb.FieldDescriptorProto_TYPE_STRING
	responseField1Label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	responseMsg := &descriptorpb.DescriptorProto{
		Name: &responseMsgName,
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   &responseField1Name,
				Number: &responseField1Number,
				Type:   &responseField1Type,
				Label:  &responseField1Label,
			},
		},
	}

	// Create file descriptor
	fileName := "test.proto"
	syntax := "proto3"

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:        &fileName,
		Package:     &packageName,
		Syntax:      &syntax,
		Service:     []*descriptorpb.ServiceDescriptorProto{service},
		MessageType: []*descriptorpb.DescriptorProto{requestMsg, responseMsg},
	}

	return &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{fileDesc},
	}
}

// createMultiServiceTestData creates test data with multiple services
func createMultiServiceTestData() *descriptorpb.FileDescriptorSet {
	packageName := "multi.v1"
	fileName := "multi.proto"
	syntax := "proto3"

	// Create two services
	service1Name := "UserService"
	service2Name := "OrderService"

	method1Name := "GetUser"
	method2Name := "GetOrder"

	inputType1 := ".multi.v1.GetUserRequest"
	outputType1 := ".multi.v1.GetUserResponse"
	inputType2 := ".multi.v1.GetOrderRequest"
	outputType2 := ".multi.v1.GetOrderResponse"

	service1 := &descriptorpb.ServiceDescriptorProto{
		Name: &service1Name,
		Method: []*descriptorpb.MethodDescriptorProto{
			{
				Name:       &method1Name,
				InputType:  &inputType1,
				OutputType: &outputType1,
			},
		},
	}

	service2 := &descriptorpb.ServiceDescriptorProto{
		Name: &service2Name,
		Method: []*descriptorpb.MethodDescriptorProto{
			{
				Name:       &method2Name,
				InputType:  &inputType2,
				OutputType: &outputType2,
			},
		},
	}

	// Create message descriptors
	getUserReqName := "GetUserRequest"
	getUserReqIdField := "id"
	getUserReqIdNumber := int32(1)
	getUserReqIdType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	getUserReqIdLabel := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	getUserRespName := "GetUserResponse"
	getUserRespNameField := "name"
	getUserRespNameNumber := int32(1)
	getUserRespNameType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	getUserRespNameLabel := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	getOrderReqName := "GetOrderRequest"
	getOrderReqIdField := "id"
	getOrderReqIdNumber := int32(1)
	getOrderReqIdType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	getOrderReqIdLabel := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	getOrderRespName := "GetOrderResponse"
	getOrderRespStatusField := "status"
	getOrderRespStatusNumber := int32(1)
	getOrderRespStatusType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	getOrderRespStatusLabel := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    &fileName,
		Package: &packageName,
		Syntax:  &syntax,
		Service: []*descriptorpb.ServiceDescriptorProto{service1, service2},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: &getUserReqName,
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   &getUserReqIdField,
						Number: &getUserReqIdNumber,
						Type:   &getUserReqIdType,
						Label:  &getUserReqIdLabel,
					},
				},
			},
			{
				Name: &getUserRespName,
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   &getUserRespNameField,
						Number: &getUserRespNameNumber,
						Type:   &getUserRespNameType,
						Label:  &getUserRespNameLabel,
					},
				},
			},
			{
				Name: &getOrderReqName,
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   &getOrderReqIdField,
						Number: &getOrderReqIdNumber,
						Type:   &getOrderReqIdType,
						Label:  &getOrderReqIdLabel,
					},
				},
			},
			{
				Name: &getOrderRespName,
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   &getOrderRespStatusField,
						Number: &getOrderRespStatusNumber,
						Type:   &getOrderRespStatusType,
						Label:  &getOrderRespStatusLabel,
					},
				},
			},
		},
	}

	return &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{fileDesc},
	}
}

// TestNew tests creating a new empty registry
func TestNew(t *testing.T) {
	registry := New()

	if registry == nil {
		t.Fatal("New() returned nil registry")
	}

	stats := registry.GetStats()
	if stats.FileCount != 0 {
		t.Errorf("Expected 0 files, got %d", stats.FileCount)
	}
	if stats.ServiceCount != 0 {
		t.Errorf("Expected 0 services, got %d", stats.ServiceCount)
	}
	if stats.MessageCount != 0 {
		t.Errorf("Expected 0 messages, got %d", stats.MessageCount)
	}
}

// TestRegister tests service registration with valid descriptor set
func TestRegister(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	err := registry.Register(fds)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	stats := registry.GetStats()
	if stats.FileCount != 1 {
		t.Errorf("Expected 1 file, got %d", stats.FileCount)
	}
	if stats.ServiceCount != 1 {
		t.Errorf("Expected 1 service, got %d", stats.ServiceCount)
	}
	if stats.MessageCount != 2 {
		t.Errorf("Expected 2 messages, got %d", stats.MessageCount)
	}
}

// TestRegister_MultipleServices tests registering multiple services
func TestRegister_MultipleServices(t *testing.T) {
	registry := New()
	fds := createMultiServiceTestData()

	err := registry.Register(fds)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	stats := registry.GetStats()
	if stats.ServiceCount != 2 {
		t.Errorf("Expected 2 services, got %d", stats.ServiceCount)
	}
	if stats.MessageCount != 4 {
		t.Errorf("Expected 4 messages, got %d", stats.MessageCount)
	}
}

// TestRegister_NilDescriptorSet tests that Register panics on nil descriptor set
// This is expected behavior as passing nil is a programming error
func TestRegister_NilDescriptorSet(t *testing.T) {
	registry := New()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil descriptor set, got nil")
		}
	}()

	_ = registry.Register(nil)
}

// TestRegister_EmptyDescriptorSet tests registering empty descriptor set
// Empty file lists are valid but result in no services being registered
func TestRegister_EmptyDescriptorSet(t *testing.T) {
	registry := New()
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{},
	}

	// Empty descriptor set should succeed but register nothing
	err := registry.Register(fds)
	if err != nil {
		t.Errorf("Register with empty descriptor set failed: %v", err)
	}

	stats := registry.GetStats()
	if stats.ServiceCount != 0 {
		t.Errorf("Expected 0 services for empty descriptor set, got %d", stats.ServiceCount)
	}
}

// TestListServices tests listing all registered services
func TestListServices(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	services := registry.ListServices()
	if len(services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(services))
	}

	svc := services[0]
	if svc.Name != "test.v1.TestService" {
		t.Errorf("Expected service name 'test.v1.TestService', got '%s'", svc.Name)
	}
	if svc.Package != "test.v1" {
		t.Errorf("Expected package 'test.v1', got '%s'", svc.Package)
	}
	if len(svc.Methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(svc.Methods))
	}

	if len(svc.Methods) > 0 {
		method := svc.Methods[0]
		if method.Name != "TestMethod" {
			t.Errorf("Expected method name 'TestMethod', got '%s'", method.Name)
		}
		if method.InputType != "test.v1.TestRequest" {
			t.Errorf("Expected input type 'test.v1.TestRequest', got '%s'", method.InputType)
		}
		if method.OutputType != "test.v1.TestResponse" {
			t.Errorf("Expected output type 'test.v1.TestResponse', got '%s'", method.OutputType)
		}
	}
}

// TestListServices_Empty tests listing services from empty registry
func TestListServices_Empty(t *testing.T) {
	registry := New()

	services := registry.ListServices()
	if len(services) != 0 {
		t.Errorf("Expected 0 services for empty registry, got %d", len(services))
	}
}

// TestListServices_Multiple tests listing multiple services
func TestListServices_Multiple(t *testing.T) {
	registry := New()
	fds := createMultiServiceTestData()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	services := registry.ListServices()
	if len(services) != 2 {
		t.Fatalf("Expected 2 services, got %d", len(services))
	}

	foundUserService := false
	foundOrderService := false

	for _, svc := range services {
		if svc.Name == "multi.v1.UserService" {
			foundUserService = true
		}
		if svc.Name == "multi.v1.OrderService" {
			foundOrderService = true
		}
	}

	if !foundUserService {
		t.Error("Expected to find UserService")
	}
	if !foundOrderService {
		t.Error("Expected to find OrderService")
	}
}

// TestGetService tests retrieving a service by name
func TestGetService(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	svc, err := registry.GetService("test.v1.TestService")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}

	if svc == nil {
		t.Fatal("Expected service descriptor, got nil")
	}

	if svc.GetFullyQualifiedName() != "test.v1.TestService" {
		t.Errorf("Expected service name 'test.v1.TestService', got '%s'", svc.GetFullyQualifiedName())
	}
}

// TestGetService_NotFound tests error handling for non-existent service
func TestGetService_NotFound(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err := registry.GetService("nonexistent.Service")
	if err == nil {
		t.Error("Expected error for non-existent service, got nil")
	}

	expectedMsg := "service not found: nonexistent.Service"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestGetService_EmptyRegistry tests getting service from empty registry
func TestGetService_EmptyRegistry(t *testing.T) {
	registry := New()

	_, err := registry.GetService("test.v1.TestService")
	if err == nil {
		t.Error("Expected error for empty registry, got nil")
	}
}

// TestClear tests clearing the registry
func TestClear(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify services are loaded
	stats := registry.GetStats()
	if stats.ServiceCount == 0 {
		t.Fatal("Expected services to be loaded before Clear")
	}

	// Clear the registry
	registry.Clear()

	// Verify everything is cleared
	statsAfter := registry.GetStats()
	if statsAfter.FileCount != 0 {
		t.Errorf("Expected 0 files after Clear, got %d", statsAfter.FileCount)
	}
	if statsAfter.ServiceCount != 0 {
		t.Errorf("Expected 0 services after Clear, got %d", statsAfter.ServiceCount)
	}
	if statsAfter.MessageCount != 0 {
		t.Errorf("Expected 0 messages after Clear, got %d", statsAfter.MessageCount)
	}

	// Verify ListServices returns empty
	services := registry.ListServices()
	if len(services) != 0 {
		t.Errorf("Expected 0 services after Clear, got %d", len(services))
	}

	// Verify GetService returns error
	_, err := registry.GetService("test.v1.TestService")
	if err == nil {
		t.Error("Expected error getting service after Clear, got nil")
	}
}

// TestRegister_Duplicate tests duplicate registration handling
func TestRegister_Duplicate(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	// First registration
	if err := registry.Register(fds); err != nil {
		t.Fatalf("First Register failed: %v", err)
	}

	statsAfterFirst := registry.GetStats()

	// Second registration (should overwrite)
	if err := registry.Register(fds); err != nil {
		t.Fatalf("Second Register failed: %v", err)
	}

	statsAfterSecond := registry.GetStats()

	// Stats should remain the same (overwrite, not add)
	if statsAfterSecond.FileCount != statsAfterFirst.FileCount {
		t.Errorf("Expected file count to remain %d, got %d", statsAfterFirst.FileCount, statsAfterSecond.FileCount)
	}
	if statsAfterSecond.ServiceCount != statsAfterFirst.ServiceCount {
		t.Errorf("Expected service count to remain %d, got %d", statsAfterFirst.ServiceCount, statsAfterSecond.ServiceCount)
	}
	if statsAfterSecond.MessageCount != statsAfterFirst.MessageCount {
		t.Errorf("Expected message count to remain %d, got %d", statsAfterFirst.MessageCount, statsAfterSecond.MessageCount)
	}
}

// TestHasService tests checking if a service is registered
func TestHasService(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	// Before registration
	if registry.HasService("test.v1.TestService") {
		t.Error("Expected HasService to return false before registration")
	}

	// Register service
	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// After registration
	if !registry.HasService("test.v1.TestService") {
		t.Error("Expected HasService to return true after registration")
	}

	// Non-existent service
	if registry.HasService("nonexistent.Service") {
		t.Error("Expected HasService to return false for non-existent service")
	}
}

// TestGetMethodDescriptor tests retrieving method descriptors
func TestGetMethodDescriptor(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	method, err := registry.GetMethodDescriptor("test.v1.TestService", "TestMethod")
	if err != nil {
		t.Fatalf("GetMethodDescriptor failed: %v", err)
	}

	if method == nil {
		t.Fatal("Expected method descriptor, got nil")
	}

	if method.GetName() != "TestMethod" {
		t.Errorf("Expected method name 'TestMethod', got '%s'", method.GetName())
	}
}

// TestGetMethodDescriptor_ServiceNotFound tests error when service doesn't exist
func TestGetMethodDescriptor_ServiceNotFound(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err := registry.GetMethodDescriptor("nonexistent.Service", "TestMethod")
	if err == nil {
		t.Error("Expected error for non-existent service, got nil")
	}
}

// TestGetMethodDescriptor_MethodNotFound tests error when method doesn't exist
func TestGetMethodDescriptor_MethodNotFound(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err := registry.GetMethodDescriptor("test.v1.TestService", "NonExistentMethod")
	if err == nil {
		t.Error("Expected error for non-existent method, got nil")
	}
}

// TestGetMessageDescriptor tests retrieving message descriptors
func TestGetMessageDescriptor(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	msg, err := registry.GetMessageDescriptor("test.v1.TestRequest")
	if err != nil {
		t.Fatalf("GetMessageDescriptor failed: %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message descriptor, got nil")
	}

	if msg.GetName() != "TestRequest" {
		t.Errorf("Expected message name 'TestRequest', got '%s'", msg.GetName())
	}
}

// TestGetMessageDescriptor_NotFound tests error when message doesn't exist
func TestGetMessageDescriptor_NotFound(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err := registry.GetMessageDescriptor("nonexistent.Message")
	if err == nil {
		t.Error("Expected error for non-existent message, got nil")
	}
}

// TestGetServiceSchema tests retrieving service schema with message schemas
func TestGetServiceSchema(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	info, schemas, err := registry.GetServiceSchema("test.v1.TestService")
	if err != nil {
		t.Fatalf("GetServiceSchema failed: %v", err)
	}

	if info == nil {
		t.Fatal("Expected service info, got nil")
	}

	if info.Name != "test.v1.TestService" {
		t.Errorf("Expected service name 'test.v1.TestService', got '%s'", info.Name)
	}

	if len(schemas) == 0 {
		t.Error("Expected message schemas, got empty map")
	}

	// Verify schemas for input and output types exist
	if _, exists := schemas["test.v1.TestRequest"]; !exists {
		t.Error("Expected schema for TestRequest")
	}
	if _, exists := schemas["test.v1.TestResponse"]; !exists {
		t.Error("Expected schema for TestResponse")
	}
}

// TestGetServiceSchema_NotFound tests error when service doesn't exist
func TestGetServiceSchema_NotFound(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, _, err := registry.GetServiceSchema("nonexistent.Service")
	if err == nil {
		t.Error("Expected error for non-existent service, got nil")
	}
}

// TestValidateDescriptors tests descriptor validation
func TestValidateDescriptors(t *testing.T) {
	tests := []struct {
		name    string
		fds     *descriptorpb.FileDescriptorSet
		wantErr bool
	}{
		{
			name:    "nil descriptor set",
			fds:     nil,
			wantErr: true,
		},
		{
			name: "empty descriptor set",
			fds: &descriptorpb.FileDescriptorSet{
				File: []*descriptorpb.FileDescriptorProto{},
			},
			wantErr: true,
		},
		{
			name:    "valid descriptor set",
			fds:     createTestFileDescriptorSet(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescriptors(tt.fds)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescriptors() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestClone tests cloning a registry
func TestClone(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Clone the registry
	clone := registry.Clone()

	// Verify clone has same stats
	originalStats := registry.GetStats()
	cloneStats := clone.GetStats()

	if cloneStats.FileCount != originalStats.FileCount {
		t.Errorf("Expected clone file count %d, got %d", originalStats.FileCount, cloneStats.FileCount)
	}
	if cloneStats.ServiceCount != originalStats.ServiceCount {
		t.Errorf("Expected clone service count %d, got %d", originalStats.ServiceCount, cloneStats.ServiceCount)
	}
	if cloneStats.MessageCount != originalStats.MessageCount {
		t.Errorf("Expected clone message count %d, got %d", originalStats.MessageCount, cloneStats.MessageCount)
	}

	// Verify clone has same services
	if !clone.HasService("test.v1.TestService") {
		t.Error("Expected clone to have TestService")
	}

	// Modify original and verify clone is unaffected
	registry.Clear()

	if registry.GetStats().ServiceCount != 0 {
		t.Error("Expected original to be cleared")
	}

	if clone.GetStats().ServiceCount == 0 {
		t.Error("Expected clone to retain services after original is cleared")
	}
}

// TestMarshalUnmarshalBinary tests binary serialization
func TestMarshalUnmarshalBinary(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Marshal to binary
	data, err := registry.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Expected non-empty binary data")
	}

	// Create new registry and unmarshal
	registry2 := New()
	if err := registry2.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}

	// Verify unmarshaled registry has same data
	stats1 := registry.GetStats()
	stats2 := registry2.GetStats()

	if stats2.FileCount != stats1.FileCount {
		t.Errorf("Expected file count %d, got %d", stats1.FileCount, stats2.FileCount)
	}
	if stats2.ServiceCount != stats1.ServiceCount {
		t.Errorf("Expected service count %d, got %d", stats1.ServiceCount, stats2.ServiceCount)
	}
	if stats2.MessageCount != stats1.MessageCount {
		t.Errorf("Expected message count %d, got %d", stats1.MessageCount, stats2.MessageCount)
	}

	if !registry2.HasService("test.v1.TestService") {
		t.Error("Expected unmarshaled registry to have TestService")
	}
}

// TestConcurrentAccess tests thread-safe concurrent access
func TestConcurrentAccess(t *testing.T) {
	registry := New()
	fds := createTestFileDescriptorSet()

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Run concurrent operations
	done := make(chan bool)

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				registry.ListServices()
				registry.GetStats()
				registry.HasService("test.v1.TestService")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify registry is still consistent
	services := registry.ListServices()
	if len(services) != 1 {
		t.Errorf("Expected 1 service after concurrent access, got %d", len(services))
	}
}

// TestParseError tests the ParseError type
func TestParseError(t *testing.T) {
	innerErr := &ParseError{
		File:    "test.proto",
		Message: "syntax error",
	}

	err := &ParseError{
		File:    "wrapper.proto",
		Message: "failed to parse",
		Cause:   innerErr,
	}

	errMsg := err.Error()
	if errMsg != "parse error in wrapper.proto: failed to parse" {
		t.Errorf("Unexpected error message: %s", errMsg)
	}

	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Error("Unwrap did not return inner error")
	}
}

// TestUnmarshalBinary_InvalidData tests error handling for invalid binary data
func TestUnmarshalBinary_InvalidData(t *testing.T) {
	registry := New()

	// Try to unmarshal invalid data
	invalidData := []byte{0x00, 0x01, 0x02, 0x03}
	err := registry.UnmarshalBinary(invalidData)
	if err == nil {
		t.Error("Expected error for invalid binary data, got nil")
	}
}

// TestValidateDescriptors_FileWithoutName tests validation catches file without name
func TestValidateDescriptors_FileWithoutName(t *testing.T) {
	emptyFileName := ""
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name: &emptyFileName,
			},
		},
	}

	err := ValidateDescriptors(fds)
	if err == nil {
		t.Error("Expected error for file descriptor with empty name, got nil")
	}
}

// TestValidateDescriptors_DuplicateFiles tests validation catches duplicate files
func TestValidateDescriptors_DuplicateFiles(t *testing.T) {
	fileName := "test.proto"
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name: &fileName,
			},
			{
				Name: &fileName,
			},
		},
	}

	err := ValidateDescriptors(fds)
	if err == nil {
		t.Error("Expected error for duplicate file descriptors, got nil")
	}
}

// TestNestedMessages tests handling of nested message types
func TestNestedMessages(t *testing.T) {
	registry := New()

	packageName := "nested.v1"
	fileName := "nested.proto"
	syntax := "proto3"

	// Create nested message structure
	innerMsgName := "Inner"
	innerFieldName := "value"
	innerFieldNumber := int32(1)
	innerFieldType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	innerFieldLabel := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	innerMsg := &descriptorpb.DescriptorProto{
		Name: &innerMsgName,
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   &innerFieldName,
				Number: &innerFieldNumber,
				Type:   &innerFieldType,
				Label:  &innerFieldLabel,
			},
		},
	}

	outerMsgName := "Outer"
	outerFieldName := "inner"
	outerFieldNumber := int32(1)
	outerFieldType := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	outerFieldLabel := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	outerFieldTypeName := ".nested.v1.Outer.Inner"

	outerMsg := &descriptorpb.DescriptorProto{
		Name: &outerMsgName,
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:     &outerFieldName,
				Number:   &outerFieldNumber,
				Type:     &outerFieldType,
				Label:    &outerFieldLabel,
				TypeName: &outerFieldTypeName,
			},
		},
		NestedType: []*descriptorpb.DescriptorProto{innerMsg},
	}

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:        &fileName,
		Package:     &packageName,
		Syntax:      &syntax,
		MessageType: []*descriptorpb.DescriptorProto{outerMsg},
	}

	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{fileDesc},
	}

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify both outer and nested message are indexed
	stats := registry.GetStats()
	if stats.MessageCount != 2 {
		t.Errorf("Expected 2 messages (outer + nested), got %d", stats.MessageCount)
	}

	// Verify we can retrieve the nested message
	_, err := registry.GetMessageDescriptor("nested.v1.Outer.Inner")
	if err != nil {
		t.Errorf("Failed to get nested message: %v", err)
	}
}

// TestMethodStreaming tests detection of streaming methods
func TestMethodStreaming(t *testing.T) {
	registry := New()

	packageName := "stream.v1"
	fileName := "stream.proto"
	syntax := "proto3"
	serviceName := "StreamService"

	inputType := ".stream.v1.StreamRequest"
	outputType := ".stream.v1.StreamResponse"

	clientStreamingTrue := true
	serverStreamingTrue := true
	clientStreamingFalse := false
	serverStreamingFalse := false

	unaryMethod := "UnaryMethod"
	clientStreamMethod := "ClientStreamMethod"
	serverStreamMethod := "ServerStreamMethod"
	bidiStreamMethod := "BidiStreamMethod"

	service := &descriptorpb.ServiceDescriptorProto{
		Name: &serviceName,
		Method: []*descriptorpb.MethodDescriptorProto{
			{
				Name:            &unaryMethod,
				InputType:       &inputType,
				OutputType:      &outputType,
				ClientStreaming: &clientStreamingFalse,
				ServerStreaming: &serverStreamingFalse,
			},
			{
				Name:            &clientStreamMethod,
				InputType:       &inputType,
				OutputType:      &outputType,
				ClientStreaming: &clientStreamingTrue,
				ServerStreaming: &serverStreamingFalse,
			},
			{
				Name:            &serverStreamMethod,
				InputType:       &inputType,
				OutputType:      &outputType,
				ClientStreaming: &clientStreamingFalse,
				ServerStreaming: &serverStreamingTrue,
			},
			{
				Name:            &bidiStreamMethod,
				InputType:       &inputType,
				OutputType:      &outputType,
				ClientStreaming: &clientStreamingTrue,
				ServerStreaming: &serverStreamingTrue,
			},
		},
	}

	reqMsgName := "StreamRequest"
	reqFieldName := "data"
	reqFieldNumber := int32(1)
	reqFieldType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	reqFieldLabel := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	respMsgName := "StreamResponse"
	respFieldName := "result"
	respFieldNumber := int32(1)
	respFieldType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	respFieldLabel := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:    &fileName,
		Package: &packageName,
		Syntax:  &syntax,
		Service: []*descriptorpb.ServiceDescriptorProto{service},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: &reqMsgName,
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   &reqFieldName,
						Number: &reqFieldNumber,
						Type:   &reqFieldType,
						Label:  &reqFieldLabel,
					},
				},
			},
			{
				Name: &respMsgName,
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   &respFieldName,
						Number: &respFieldNumber,
						Type:   &respFieldType,
						Label:  &respFieldLabel,
					},
				},
			},
		},
	}

	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{fileDesc},
	}

	if err := registry.Register(fds); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// List services and verify streaming flags
	services := registry.ListServices()
	if len(services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(services))
	}

	svc := services[0]
	if len(svc.Methods) != 4 {
		t.Fatalf("Expected 4 methods, got %d", len(svc.Methods))
	}

	// Check each method type
	methodTests := map[string]struct {
		wantClientStreaming bool
		wantServerStreaming bool
	}{
		"UnaryMethod":        {false, false},
		"ClientStreamMethod": {true, false},
		"ServerStreamMethod": {false, true},
		"BidiStreamMethod":   {true, true},
	}

	for _, method := range svc.Methods {
		expected, ok := methodTests[method.Name]
		if !ok {
			t.Errorf("Unexpected method: %s", method.Name)
			continue
		}

		if method.ClientStreaming != expected.wantClientStreaming {
			t.Errorf("Method %s: expected ClientStreaming=%v, got %v",
				method.Name, expected.wantClientStreaming, method.ClientStreaming)
		}

		if method.ServerStreaming != expected.wantServerStreaming {
			t.Errorf("Method %s: expected ServerStreaming=%v, got %v",
				method.Name, expected.wantServerStreaming, method.ServerStreaming)
		}
	}
}
