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

	CreateChatCompletionStream(
		ctx context.Context,
		request Request,
		mode Mode,
		schema *Schema,
	) (<-chan string, error)
}
