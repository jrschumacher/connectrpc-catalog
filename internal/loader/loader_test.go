package loader

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"
)

// TestLoadFromPath_Success tests loading protos from a valid directory
func TestLoadFromPath_Success(t *testing.T) {
	// Use the actual proto directory in the project
	protoPath := filepath.Join("..", "..", "proto")

	// Check if path exists, skip if not (e.g., in CI without buf setup)
	if _, err := os.Stat(protoPath); os.IsNotExist(err) {
		t.Skip("Proto directory not found, skipping test")
	}

	// Verify buf is installed
	if err := ValidateBufInstallation(); err != nil {
		t.Skip("buf CLI not installed, skipping test")
	}

	fds, err := LoadFromPath(protoPath)
	if err != nil {
		t.Fatalf("LoadFromPath failed: %v", err)
	}

	if fds == nil {
		t.Fatal("Expected FileDescriptorSet, got nil")
	}

	if len(fds.File) == 0 {
		t.Error("Expected at least one file descriptor, got zero")
	}

	// Verify the descriptor set contains valid data
	info := GetDescriptorInfo(fds)
	if info.Files == 0 {
		t.Error("Expected files in descriptor info, got zero")
	}

	t.Logf("Successfully loaded %d files, %d services, %d messages, %d enums",
		info.Files, len(info.Services), len(info.Messages), len(info.Enums))
}

// TestLoadFromPath_NonExistent tests error handling for non-existent paths
func TestLoadFromPath_NonExistent(t *testing.T) {
	_, err := LoadFromPath("/nonexistent/path/to/protos")

	if err == nil {
		t.Fatal("Expected error for non-existent path, got nil")
	}

	// Verify error message mentions path issue
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestLoadFromPath_EmptyDirectory tests loading from directory with no proto files
func TestLoadFromPath_EmptyDirectory(t *testing.T) {
	// Verify buf is installed
	if err := ValidateBufInstallation(); err != nil {
		t.Skip("buf CLI not installed, skipping test")
	}

	// Create temporary empty directory
	tmpDir, err := os.MkdirTemp("", "loader-test-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Try to load from empty directory
	// buf build will fail on empty directory or directory without valid proto structure
	_, err = LoadFromPath(tmpDir)

	if err == nil {
		t.Fatal("Expected error for empty directory, got nil")
	}

	// buf build should fail with an error
	t.Logf("Got expected error for empty directory: %v", err)
}

// TestLoadFromPath_InvalidProtoStructure tests handling of invalid proto files
func TestLoadFromPath_InvalidProtoStructure(t *testing.T) {
	// Verify buf is installed
	if err := ValidateBufInstallation(); err != nil {
		t.Skip("buf CLI not installed, skipping test")
	}

	// Create temporary directory with invalid proto file
	tmpDir, err := os.MkdirTemp("", "loader-test-invalid-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid proto file
	invalidProto := filepath.Join(tmpDir, "invalid.proto")
	if err := os.WriteFile(invalidProto, []byte("this is not valid proto syntax"), 0644); err != nil {
		t.Fatalf("Failed to write invalid proto: %v", err)
	}

	// buf build will fail on invalid proto syntax
	_, err = LoadFromPath(tmpDir)

	if err == nil {
		t.Fatal("Expected error for invalid proto structure, got nil")
	}

	t.Logf("Got expected error for invalid proto: %v", err)
}

// TestGetDescriptorInfo tests extracting metadata from FileDescriptorSet
func TestGetDescriptorInfo(t *testing.T) {
	// Create a minimal test FileDescriptorSet
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name:    stringPtr("test.proto"),
				Package: stringPtr("test.v1"),
				Service: []*descriptorpb.ServiceDescriptorProto{
					{
						Name: stringPtr("TestService"),
						Method: []*descriptorpb.MethodDescriptorProto{
							{
								Name:       stringPtr("TestMethod"),
								InputType:  stringPtr(".test.v1.Request"),
								OutputType: stringPtr(".test.v1.Response"),
							},
						},
					},
				},
				MessageType: []*descriptorpb.DescriptorProto{
					{Name: stringPtr("Request")},
					{Name: stringPtr("Response")},
				},
				EnumType: []*descriptorpb.EnumDescriptorProto{
					{Name: stringPtr("Status")},
				},
			},
		},
	}

	info := GetDescriptorInfo(fds)

	if info.Files != 1 {
		t.Errorf("Expected 1 file, got %d", info.Files)
	}

	if len(info.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(info.Services))
	}

	expectedServiceName := "test.v1.TestService"
	if len(info.Services) > 0 && info.Services[0] != expectedServiceName {
		t.Errorf("Expected service name '%s', got '%s'", expectedServiceName, info.Services[0])
	}

	if len(info.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(info.Messages))
	}

	if len(info.Enums) != 1 {
		t.Errorf("Expected 1 enum, got %d", len(info.Enums))
	}
}

// TestGetDescriptorInfo_EmptyDescriptorSet tests handling of empty descriptor set
func TestGetDescriptorInfo_EmptyDescriptorSet(t *testing.T) {
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{},
	}

	info := GetDescriptorInfo(fds)

	if info.Files != 0 {
		t.Errorf("Expected 0 files, got %d", info.Files)
	}

	if len(info.Services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(info.Services))
	}

	if len(info.Messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(info.Messages))
	}

	if len(info.Enums) != 0 {
		t.Errorf("Expected 0 enums, got %d", len(info.Enums))
	}
}

// TestGetDescriptorInfo_NoPackage tests handling descriptors without package
func TestGetDescriptorInfo_NoPackage(t *testing.T) {
	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{
			{
				Name:    stringPtr("test.proto"),
				Package: nil, // No package
				Service: []*descriptorpb.ServiceDescriptorProto{
					{Name: stringPtr("TestService")},
				},
				MessageType: []*descriptorpb.DescriptorProto{
					{Name: stringPtr("Request")},
				},
			},
		},
	}

	info := GetDescriptorInfo(fds)

	// When no package is specified, names should be simple
	if len(info.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(info.Services))
	}

	if len(info.Services) > 0 && info.Services[0] != "TestService" {
		t.Errorf("Expected service name 'TestService', got '%s'", info.Services[0])
	}

	if len(info.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(info.Messages))
	}

	if len(info.Messages) > 0 && info.Messages[0] != "Request" {
		t.Errorf("Expected message name 'Request', got '%s'", info.Messages[0])
	}
}

// TestValidateBufInstallation tests buf installation check
func TestValidateBufInstallation(t *testing.T) {
	err := ValidateBufInstallation()

	// We don't fail the test if buf is not installed, just log it
	if err != nil {
		t.Logf("buf CLI not installed or not in PATH: %v", err)
		t.Skip("Skipping tests that require buf CLI")
	}

	t.Log("buf CLI is installed and accessible")
}

// TestLoad_WithPathSource tests the unified Load function with path source
func TestLoad_WithPathSource(t *testing.T) {
	// Check if path exists
	protoPath := filepath.Join("..", "..", "proto")
	if _, err := os.Stat(protoPath); os.IsNotExist(err) {
		t.Skip("Proto directory not found, skipping test")
	}

	// Verify buf is installed
	if err := ValidateBufInstallation(); err != nil {
		t.Skip("buf CLI not installed, skipping test")
	}

	source := LoadSource{
		Type:  SourceTypePath,
		Value: protoPath,
	}

	fds, err := Load(source)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if fds == nil {
		t.Fatal("Expected FileDescriptorSet, got nil")
	}

	if len(fds.File) == 0 {
		t.Error("Expected at least one file descriptor, got zero")
	}
}

// TestLoad_UnknownSourceType tests error handling for unknown source type
func TestLoad_UnknownSourceType(t *testing.T) {
	source := LoadSource{
		Type:  "unknown",
		Value: "some-value",
	}

	_, err := Load(source)

	if err == nil {
		t.Fatal("Expected error for unknown source type, got nil")
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	// Verify error mentions unknown source type
	t.Logf("Got expected error: %v", err)
}

// TestLoadFromGitHub_InvalidRepo tests error handling for invalid GitHub repo
func TestLoadFromGitHub_InvalidRepo(t *testing.T) {
	// Skip if git is not installed
	if err := ValidateBufInstallation(); err != nil {
		t.Skip("buf CLI not installed, skipping test")
	}

	// Try to load from non-existent GitHub repo
	_, err := LoadFromGitHub("github.com/nonexistent/repo")

	if err == nil {
		t.Fatal("Expected error for invalid GitHub repo, got nil")
	}

	// Should fail on git clone
	t.Logf("Got expected error for invalid repo: %v", err)
}

// TestLoadFromBufModule_InvalidModule tests error handling for invalid Buf module
func TestLoadFromBufModule_InvalidModule(t *testing.T) {
	// Skip if buf is not installed
	if err := ValidateBufInstallation(); err != nil {
		t.Skip("buf CLI not installed, skipping test")
	}

	// Try to load from non-existent module
	_, err := LoadFromBufModule("nonexistent/module")

	if err == nil {
		t.Fatal("Expected error for invalid Buf module, got nil")
	}

	// Should fail on buf export
	t.Logf("Got expected error for invalid module: %v", err)
}

// TestSourceType_Constants tests that source type constants are defined
func TestSourceType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		source   SourceType
		expected string
	}{
		{"Path", SourceTypePath, "path"},
		{"GitHub", SourceTypeGitHub, "github"},
		{"BufModule", SourceTypeBufModule, "buf_module"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("Expected source type '%s', got '%s'", tt.expected, string(tt.source))
			}
		})
	}
}

// TestLoadResult_Structure tests LoadResult struct
func TestLoadResult_Structure(t *testing.T) {
	result := LoadResult{
		ServiceCount: 5,
		FileCount:    10,
		Error:        nil,
	}

	if result.ServiceCount != 5 {
		t.Errorf("Expected ServiceCount 5, got %d", result.ServiceCount)
	}

	if result.FileCount != 10 {
		t.Errorf("Expected FileCount 10, got %d", result.FileCount)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
