package instructor

import (
	"context"

	"github.com/madebywelch/anthropic-go/v2/pkg/anthropic"
	"github.com/sashabaranov/go-openai"
)

type Instructor[T any] struct {
	Client     Client[T]
	Provider   Provider
	Mode       Mode
	MaxRetries int
}

func FromOpenAI[T any](client *openai.Client, opts ...Options) (*Instructor[T], error) {
	options := mergeOptions(opts...)

	openaiClient, err := NewOpenAIClient[T](client, *options.Mode, *options.MaxRetries)
	if err != nil {
		return nil, err
	}

	i := &Instructor[T]{
		Client:     openaiClient,
		Provider:   OpenAI,
		Mode:       *options.Mode,
		MaxRetries: *options.MaxRetries,
	}
	return i, nil
}

func FromAnthropic[T any](cli *anthropic.Client) (*Instructor[T], error) {
	panic("not implemented")
}

// CreateChatCompletion implements Client.
func (i *Instructor[T]) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest) (*T, error) {
	return i.Client.CreateChatCompletion(ctx, request)
}
