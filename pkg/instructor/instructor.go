package instructor

import (
	"context"

	"github.com/madebywelch/anthropic-go/v2/pkg/anthropic"
	"github.com/sashabaranov/go-openai"
)

type Instructor[T any] struct {
	Client   Client[T]
	Provider Provider
}

func FromOpenAI[T any](client *openai.Client) *Instructor[T] {
	i := &Instructor[T]{
		Client:   &OpenAIClient[T]{client: client},
		Provider: OpenAI,
	}
	return i
}

func FromAnthropic[T any](cli *anthropic.Client) *Instructor[T] {
	panic("not implemented")
}

// CreateChatCompletion implements Client.
func (i *Instructor[T]) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest, opts ...ClientOptions) (*T, error) {
	return i.Client.CreateChatCompletion(ctx, request, opts...)
}
