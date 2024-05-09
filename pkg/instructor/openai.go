package instructor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient[T any] struct {
	client     *openai.Client
	jsonSchema jsonschema.Schema
}

var _ Client[struct{}] = &OpenAIClient[struct{}]{}

func (o *OpenAIClient[T]) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest, opts ...ClientOptions) (*T, error) {

	t := new(T)
	tSchema := jsonschema.Reflect(t)

	jsonSchema, err := json.MarshalIndent(tSchema, "", "  ")
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, jsonSchema)

	msg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}

	request.Messages = insertAtFront(request.Messages, msg)

	fmt.Printf("request:\n\b%+v\n\n", request)

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest(request))
	if err != nil {
		return nil, err
	}

	text := resp.Choices[0].Message.Content

	fmt.Printf("Returned text:\n\n%s\n\n", text)

	err = json.Unmarshal([]byte(text), t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// TODO: move to utils
func insertAtFront[T any](to []T, from T) []T {
	return append([]T{from}, to...)
}
