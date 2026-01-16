package server

import (
	"google.golang.org/protobuf/types/descriptorpb"
)

// createTestFileDescriptorSet creates a minimal FileDescriptorSet for testing
// This simulates what would be loaded from a proto file
func createTestFileDescriptorSet() *descriptorpb.FileDescriptorSet {
	// Create a simple test service with one method
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

	// Create file descriptor set
	return &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{fileDesc},
	}
}
