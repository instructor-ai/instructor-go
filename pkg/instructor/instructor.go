package instructor

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/invopop/jsonschema"

	"github.com/madebywelch/anthropic-go/v2/pkg/anthropic"
	"github.com/sashabaranov/go-openai"
)

type Instructor[T any] struct {
	Client     Client
	Provider   Provider
	Mode       Mode
	MaxRetries int

	Schema    *jsonschema.Schema
	SchemaStr []byte
}

func FromOpenAI[T any](client *openai.Client, opts ...Options) (*Instructor[T], error) {

	options := mergeOptions(opts...)

	schema := jsonschema.Reflect(new(T))
	schemaStr, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	openaiClient, err := NewOpenAIClient(client, schemaStr, *options.Mode)
	if err != nil {
		return nil, err
	}

	i := &Instructor[T]{
		Client:     openaiClient,
		Provider:   OpenAI,
		Mode:       *options.Mode,
		MaxRetries: *options.MaxRetries,
		Schema:     schema,
		SchemaStr:  schemaStr,
	}
	return i, nil
}

func FromAnthropic[T any](cli *anthropic.Client) (*Instructor[T], error) {
	panic("not implemented")
}

func (i *Instructor[T]) CreateChatCompletion(ctx context.Context, request Request) (*T, error) {

	for attempt := 0; attempt < i.MaxRetries; attempt++ {

		text, err := i.Client.CreateChatCompletion(ctx, request)
		if err != nil {
			// no retry on non-marshalling/validation errors
			return nil, err
		}

		t, err := i.processResponse(text)
		if err != nil {
			// TODO:
			// add more sophisticated retry logic (send back json and parse error for model to fix).
			//
			// Currently, its just recalling with no new information
			// or attempt to fix the error with the last generated JSON
			continue
		}

		return t, nil
	}

	return nil, errors.New("hit max retry attempts")
}

func (i *Instructor[T]) processResponse(response string) (*T, error) {

	t := new(T)

	err := json.Unmarshal([]byte(response), t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
