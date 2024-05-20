package instructor

import (
	"context"
)

type Client interface {
	CreateChatCompletion(
		ctx context.Context,
		request Request,
		mode Mode,
		schema *Schema,
	) (string, error)

	// TODO: implement streaming
	// CreateChatCompletionStream(
	// 	ctx context.Context,
	// 	request ChatCompletionRequest,
	// 	opts ...ClientOptions,
	// ) (*T, error)
}
