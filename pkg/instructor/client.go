package instructor

import (
	"context"

	. "github.com/instructor-ai/instructor-go/pkg/instructor/modes"
)

type Client[T any] interface {
	CreateChatCompletion(
		ctx context.Context,
		request ChatCompletionRequest,
		opts ...ClientOptions,
	) (*T, error)

	// TODO: implement streaming
	// CreateChatCompletionStream(
	// 	ctx context.Context,
	// 	request ChatCompletionRequest,
	// 	opts ...ClientOptions,
	// ) (*T, error)
}

type ClientOptions struct {
	Mode       Mode
	MaxRetries int

	// Provider specific options:
}

func WithMode(mode Mode) ClientOptions {
	return ClientOptions{Mode: mode}
}

func WithMaxRetries(maxRetries int) ClientOptions {
	return ClientOptions{MaxRetries: maxRetries}
}
