package instructor

import (
	"context"
)

type Instructor interface {
	Provider() Provider
	Mode() Mode
	MaxRetries() int

	// Chat / Messages

	chat(
		ctx context.Context,
		request interface{},
		schema *Schema,
	) (string, interface{}, error)

	// Streaming Chat / Messages

	chatStream(
		ctx context.Context,
		request interface{},
		schema *Schema,
	) (<-chan string, error)
}
