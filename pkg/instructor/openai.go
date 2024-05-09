package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/invopop/jsonschema"
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient[T any] struct {
	client *openai.Client

	mode       Mode
	maxRetries int

	schema    *jsonschema.Schema
	schemaStr []byte
}

var _ Client[struct{}] = &OpenAIClient[struct{}]{}

func NewOpenAIClient[T any](client *openai.Client, mode Mode, maxRetries int) (*OpenAIClient[T], error) {

	schema := jsonschema.Reflect(new(T))
	schemaStr, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	o := &OpenAIClient[T]{
		client:     client,
		mode:       mode,
		maxRetries: maxRetries,
		schema:     schema,
		schemaStr:  schemaStr,
	}
	return o, nil
}

func (o *OpenAIClient[T]) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest) (*T, error) {

	for attempt := 0; attempt < o.maxRetries; attempt++ {

		text, err := o.createChatCompletionModeHandler(ctx, request)
		if err != nil {
			// don't retry marshalling + validation error
			return nil, err
		}

		println(text)

		// Process response

		t := new(T)

		err = json.Unmarshal([]byte(text), t)
		if err != nil {
			// TODO: add more sophisticated retry logic (send back json and parse error for model to fix).
			//       Now, its just retrying from scratch
			continue
		}

		return t, nil
	}

	return nil, errors.New("hit max retry attempts")
}

func (o *OpenAIClient[T]) createChatCompletionModeHandler(ctx context.Context, request ChatCompletionRequest) (string, error) {
	switch o.mode {
	case ModeToolCall:
		return o.createChatCompletionToolCall(ctx, request)
	case ModeJSON:
		return o.createChatCompletionJSON(ctx, request)
	case ModeJSONSchema:
		return o.createChatCompletionJSONSchema(ctx, request)
	default:
		return "", fmt.Errorf("mode '%s' is not supported for OpenAI", o.mode)
	}
}

func (o *OpenAIClient[T]) createChatCompletionToolCall(ctx context.Context, request ChatCompletionRequest) (string, error) {
	panic("not implemented")
}

func (o *OpenAIClient[T]) createChatCompletionJSON(ctx context.Context, request ChatCompletionRequest) (string, error) {
	message := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, o.schemaStr)

	msg := openai.ChatCompletionMessage{
		Role:    RoleSystem,
		Content: message,
	}

	request.Messages = prepend(request.Messages, msg)

	// Set JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONObject,
	}

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest(request))
	if err != nil {
		return "", err
	}

	text := resp.Choices[0].Message.Content

	return text, nil
}

func (o *OpenAIClient[T]) createChatCompletionJSONSchema(ctx context.Context, request ChatCompletionRequest) (string, error) {

	message := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, o.schemaStr)

	msg := openai.ChatCompletionMessage{
		Role:    RoleSystem,
		Content: message,
	}

	request.Messages = prepend(request.Messages, msg)

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest(request))
	if err != nil {
		return "", err
	}

	text := resp.Choices[0].Message.Content

	return text, nil
}
