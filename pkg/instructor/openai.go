package instructor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient[T any] struct {
	client *openai.Client
	mode   Mode

	schema    *jsonschema.Schema
	schemaStr []byte
}

var _ Client[struct{}] = &OpenAIClient[struct{}]{}

func NewOpenAIClient[T any](client *openai.Client, mode Mode) (*OpenAIClient[T], error) {

	schema := jsonschema.Reflect(new(T))
	schemaStr, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	o := &OpenAIClient[T]{
		client:    client,
		mode:      mode,
		schema:    schema,
		schemaStr: schemaStr,
	}
	return o, nil
}

func (o *OpenAIClient[T]) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest, opts ...Options) (*T, error) {

	switch o.mode {
	case ModeTool:
	case ModeJSON:
	case ModeJSONSchema:
	}

	t := new(T)

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
		return nil, err
	}

	text := resp.Choices[0].Message.Content

	err = json.Unmarshal([]byte(text), t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
