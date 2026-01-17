// Package elizaservice provides an internal Eliza service implementation for testing.
// This service supports all three protocols: Connect, gRPC, and gRPC-Web.
package elizaservice

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	elizav1 "github.com/opentdf/connectrpc-catalog/gen/connectrpc/eliza/v1"
)

// Handler implements the ElizaServiceHandler interface.
type Handler struct{}

// NewHandler creates a new Eliza service handler.
func NewHandler() *Handler {
	return &Handler{}
}

// Say handles the Say RPC - responds based on the input sentence.
func (h *Handler) Say(
	ctx context.Context,
	req *connect.Request[elizav1.SayRequest],
) (*connect.Response[elizav1.SayResponse], error) {
	sentence := req.Msg.GetSentence()
	if sentence == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("sentence is required"))
	}

	response := generateResponse(sentence)
	return connect.NewResponse(&elizav1.SayResponse{
		Sentence: response,
	}), nil
}

// Converse handles the bidirectional streaming Converse RPC.
func (h *Handler) Converse(
	ctx context.Context,
	stream *connect.BidiStream[elizav1.ConverseRequest, elizav1.ConverseResponse],
) error {
	for {
		req, err := stream.Receive()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		response := generateResponse(req.GetSentence())
		if err := stream.Send(&elizav1.ConverseResponse{
			Sentence: response,
		}); err != nil {
			return err
		}
	}
}

// Introduce handles the server streaming Introduce RPC.
func (h *Handler) Introduce(
	ctx context.Context,
	req *connect.Request[elizav1.IntroduceRequest],
	stream *connect.ServerStream[elizav1.IntroduceResponse],
) error {
	name := req.Msg.GetName()
	if name == "" {
		name = "stranger"
	}

	introductions := []string{
		fmt.Sprintf("Hello, %s!", name),
		"I'm Eliza, your digital therapist.",
		"I'm here to help you explore your thoughts and feelings.",
		"How are you feeling today?",
	}

	for _, intro := range introductions {
		if err := stream.Send(&elizav1.IntroduceResponse{
			Sentence: intro,
		}); err != nil {
			return err
		}
	}

	return nil
}

// generateResponse creates a response based on the input.
func generateResponse(input string) string {
	input = strings.ToLower(input)

	switch {
	case strings.Contains(input, "hello") || strings.Contains(input, "hi"):
		return "Hello! How can I help you today?"
	case strings.Contains(input, "how are you"):
		return "I'm doing well, thank you for asking!"
	case strings.Contains(input, "test"):
		return "Test received successfully!"
	case strings.Contains(input, "help"):
		return "I'm here to help. What would you like to know?"
	case strings.Contains(input, "bye") || strings.Contains(input, "goodbye"):
		return "Goodbye! Have a great day!"
	case strings.Contains(input, "feel"):
		return "Tell me more about how you're feeling."
	case strings.Contains(input, "think"):
		return "What makes you think that?"
	case strings.Contains(input, "why"):
		return "Why do you ask?"
	default:
		return fmt.Sprintf("You said: %s. Tell me more about that.", input)
	}
}
