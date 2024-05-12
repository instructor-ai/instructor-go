package instructor

import (
	"context"
)

type Client[T any] interface {
	CreateChatCompletion(
		ctx context.Context,
		request Request,
	) (string, error)

	// TODO: implement streaming
	// CreateChatCompletionStream(
	// 	ctx context.Context,
	// 	request ChatCompletionRequest,
	// 	opts ...ClientOptions,
	// ) (*T, error)
}
