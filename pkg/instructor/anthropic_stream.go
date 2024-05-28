package instructor

import (
	"context"
)

func (a *AnthropicClient) ChatStream(ctx context.Context, request interface{}, mode Mode, schema *Schema) (<-chan string, error) {
	panic("unimplemented")
}
