package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// LoadFromPath loads proto descriptors from a local filesystem path using buf build
func LoadFromPath(path string) (*descriptorpb.FileDescriptorSet, error) {
	// Verify path exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	// Create temporary file for buf build output
	tmpFile, err := os.CreateTemp("", "connectrpc-catalog-*.bin")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Run buf build to generate descriptor set
	cmd := exec.Command("buf", "build", path, "-o", tmpPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("buf build failed: %w (stderr: %s)", err, stderr.String())
	}

	// Read the generated descriptor set
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read descriptor set: %w", err)
	}

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(data, fds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal descriptor set: %w", err)
	}

	return fds, nil
}

// LoadFromGitHub loads proto descriptors from a GitHub repository
// Expected format: "github.com/owner/repo" or "github.com/owner/repo/subdir"
func LoadFromGitHub(repo string) (*descriptorpb.FileDescriptorSet, error) {
	// Create temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "connectrpc-catalog-git-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repository
	gitURL := fmt.Sprintf("https://%s.git", repo)
	cmd := exec.Command("git", "clone", "--depth", "1", gitURL, tmpDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git clone failed: %w (stderr: %s)", err, stderr.String())
	}

	// Load protos from the cloned directory
	return LoadFromPath(tmpDir)
}

// LoadFromBufModule loads proto descriptors from a Buf registry module
// Expected format: "buf.build/owner/repo" or "owner/repo"
func LoadFromBufModule(module string) (*descriptorpb.FileDescriptorSet, error) {
	// Create temporary directory for buf export
	tmpDir, err := os.MkdirTemp("", "connectrpc-catalog-buf-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Step 1: Export the module from BSR to local directory
	exportCmd := exec.Command("buf", "export", module, "-o", tmpDir)
	var exportStderr bytes.Buffer
	exportCmd.Stderr = &exportStderr

	if err := exportCmd.Run(); err != nil {
		return nil, fmt.Errorf("buf export from module failed: %w (stderr: %s)", err, exportStderr.String())
	}

	// Create temporary file for buf build output
	tmpFile, err := os.CreateTemp("", "connectrpc-catalog-buf-*.bin")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Step 2: Build descriptor set from exported protos
	buildCmd := exec.Command("buf", "build", tmpDir, "-o", tmpPath)
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr

	if err := buildCmd.Run(); err != nil {
		return nil, fmt.Errorf("buf build from exported module failed: %w (stderr: %s)", err, buildStderr.String())
	}

	// Read the generated descriptor set
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read descriptor set: %w", err)
	}

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(data, fds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal descriptor set: %w", err)
	}

	return fds, nil
}

// LoadResult contains statistics about a load operation
type LoadResult struct {
	ServiceCount int
	FileCount    int
	Error        error
}

// DescriptorInfo provides metadata about loaded descriptors
type DescriptorInfo struct {
	Files    int
	Services []string
	Messages []string
	Enums    []string
}

// GetDescriptorInfo extracts metadata from a FileDescriptorSet
func GetDescriptorInfo(fds *descriptorpb.FileDescriptorSet) DescriptorInfo {
	info := DescriptorInfo{
		Files:    len(fds.File),
		Services: make([]string, 0),
		Messages: make([]string, 0),
		Enums:    make([]string, 0),
	}

	for _, file := range fds.File {
		pkg := file.GetPackage()

		// Collect service names
		for _, svc := range file.Service {
			fullName := svc.GetName()
			if pkg != "" {
				fullName = pkg + "." + fullName
			}
			info.Services = append(info.Services, fullName)
		}

		// Collect message names
		for _, msg := range file.MessageType {
			fullName := msg.GetName()
			if pkg != "" {
				fullName = pkg + "." + fullName
			}
			info.Messages = append(info.Messages, fullName)
		}

		// Collect enum names
		for _, enum := range file.EnumType {
			fullName := enum.GetName()
			if pkg != "" {
				fullName = pkg + "." + fullName
			}
			info.Enums = append(info.Enums, fullName)
		}
	}

	return info
}

// ValidateBufInstallation checks if buf is installed and accessible
func ValidateBufInstallation() error {
	cmd := exec.Command("buf", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buf not installed or not in PATH: %w", err)
	}
	return nil
}

// SourceType represents the type of proto source
type SourceType string

const (
	SourceTypePath       SourceType = "path"
	SourceTypeGitHub     SourceType = "github"
	SourceTypeBufModule  SourceType = "buf_module"
	SourceTypeReflection SourceType = "reflection"
)

// LoadSource represents a proto source configuration
type LoadSource struct {
	Type             SourceType
	Value            string
	ReflectionOptions *ReflectionOptions // Optional, only for reflection sources
}

// Load is a unified loader that dispatches to the appropriate loader function
func Load(source LoadSource) (*descriptorpb.FileDescriptorSet, error) {
	switch source.Type {
	case SourceTypePath:
		return LoadFromPath(source.Value)
	case SourceTypeGitHub:
		return LoadFromGitHub(source.Value)
	case SourceTypeBufModule:
		return LoadFromBufModule(source.Value)
	case SourceTypeReflection:
		opts := ReflectionOptions{}
		if source.ReflectionOptions != nil {
			opts = *source.ReflectionOptions
		}
		return LoadFromReflection(source.Value, opts)
	default:
		return nil, fmt.Errorf("unknown source type: %s", source.Type)
	}
}

// ParseBufModuleJSON parses buf module metadata (for future use)
type BufModule struct {
	Name       string   `json:"name"`
	Owner      string   `json:"owner"`
	Repository string   `json:"repository"`
	Tags       []string `json:"tags"`
}

// GetBufModuleInfo retrieves module information from BSR
func GetBufModuleInfo(module string) (*BufModule, error) {
	cmd := exec.Command("buf", "registry", "module", "info", module, "--format", "json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("buf registry module info failed: %w (stderr: %s)", err, stderr.String())
	}

	var info BufModule
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return nil, fmt.Errorf("failed to parse module info: %w", err)
	}

	return &info, nil
}
