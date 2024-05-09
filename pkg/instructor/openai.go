package instructor

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client

	schemaStr []byte
	mode      Mode
}

var _ Client = &OpenAIClient{}

func NewOpenAIClient(client *openai.Client, schemaStr []byte, mode Mode) (*OpenAIClient, error) {
	o := &OpenAIClient{
		client:    client,
		mode:      mode,
		schemaStr: schemaStr,
	}
	return o, nil
}

func (o *OpenAIClient) CreateChatCompletion(ctx context.Context, request Request) (string, error) {
	return o.completionModeHandler(ctx, request)
}

func (o *OpenAIClient) completionModeHandler(ctx context.Context, request Request) (string, error) {
	switch o.mode {
	case ModeToolCall:
		return o.completionToolCall(ctx, request)
	case ModeJSON:
		return o.completionJSON(ctx, request)
	case ModeJSONSchema:
		return o.completionJSONSchema(ctx, request)
	default:
		return "", fmt.Errorf("mode '%s' is not supported for OpenAI", o.mode)
	}
}

func (o *OpenAIClient) completionToolCall(ctx context.Context, request Request) (string, error) {
	panic("not implemented")
}

func (o *OpenAIClient) completionJSON(ctx context.Context, request Request) (string, error) {
	message := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, o.schemaStr)

	msg := Message{
		Role:    RoleSystem,
		Content: message,
	}

	request.Messages = prepend(request.Messages, msg)

	// Set JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}

	resp, err := o.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	text := resp.Choices[0].Message.Content

	return text, nil
}

func (o *OpenAIClient) completionJSONSchema(ctx context.Context, request Request) (string, error) {

	message := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, o.schemaStr)

	msg := Message{
		Role:    RoleSystem,
		Content: message,
	}

	request.Messages = prepend(request.Messages, msg)

	resp, err := o.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	text := resp.Choices[0].Message.Content

	return text, nil
}
