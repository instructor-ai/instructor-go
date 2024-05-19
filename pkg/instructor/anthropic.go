package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	anthropic "github.com/liushuangls/go-anthropic/v2"
)

type AnthropicClient[T any] struct {
	Name string

	client *anthropic.Client
	schema *Schema[T]
	mode   Mode
}

var _ Client[any] = &AnthropicClient[any]{}

func NewAnthropicClient[T any](client *anthropic.Client, schema *Schema[T], mode Mode) (*AnthropicClient[T], error) {
	o := &AnthropicClient[T]{
		Name:   "Anthropic",
		client: client,
		schema: schema,
		mode:   mode,
	}
	return o, nil
}

func (a *AnthropicClient[T]) CreateChatCompletion(ctx context.Context, request Request) (string, error) {
	return a.completionModeHandler(ctx, request)
}

func (a *AnthropicClient[any]) completionModeHandler(ctx context.Context, request Request) (string, error) {
	switch a.mode {
	case ModeToolCall:
		return a.completionToolCall(ctx, request)
	case ModeJSONSchema:
		return a.completionJSONSchema(ctx, request)
	default:
		return "", fmt.Errorf("mode '%s' is not supported for %s", a.mode, a.Name)
	}
}

func (a *AnthropicClient[any]) completionToolCall(ctx context.Context, request Request) (string, error) {

	tools := []anthropic.ToolDefinition{}

	for _, function := range a.schema.Functions {
		t := anthropic.ToolDefinition{
			Name:        function.Name,
			Description: function.Description,
			InputSchema: function.Parameters,
		}
		tools = append(tools, t)
	}

	messages, err := toAnthropicMessages(&request)
	if err != nil {
		return "", err
	}

	req := anthropic.MessagesRequest{
		Model:     request.Model,
		Messages:  *messages,
		Tools:     tools,
		MaxTokens: 1000, // TODO: make configurable
	}

	resp, err := a.client.CreateMessages(ctx, req)
	if err != nil {
		return "", err
	}

	for _, c := range resp.Content {
		if c.Type != anthropic.MessagesContentTypeToolUse {
			// Skip non tool responses
			continue
		}

		toolInput, err := json.Marshal(c.Input)
		if err != nil {
			return "", err
		}
		// TODO: handle more than 1 tool use
		return string(toolInput), nil
	}

	return "", errors.New("not implemented")

}

func (a *AnthropicClient[any]) completionJSONSchema(ctx context.Context, request Request) (string, error) {

	system := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself.
`, a.schema.String)

	messages, err := toAnthropicMessages(&request)
	if err != nil {
		return "", err
	}

	req := anthropic.MessagesRequest{
		Model:     request.Model,
		System:    system,
		Messages:  *messages,
		MaxTokens: 1000, // TODO: make configurable
	}

	resp, err := a.client.CreateMessages(ctx, req)
	if err != nil {
		return "", err
	}

	text := resp.Content[0].Text

	return *text, nil
}

func toAnthropicMessages(request *Request) (*[]anthropic.Message, error) {

	messages := make([]anthropic.Message, len(request.Messages))

	for i, msg := range request.Messages {
		switch msg.Role {
		case RoleUser:
			if msg.Content != "" {
				messages[i] = anthropic.NewUserTextMessage(msg.Content)
			} else if msg.MultiContent != nil {
				content := make([]anthropic.MessageContent, len(msg.MultiContent))
				for j, m := range msg.MultiContent {
					switch m.Type {
					case ChatMessagePartTypeText:
						content[j] = anthropic.NewTextMessageContent(m.Text)
					case ChatMessagePartTypeImageURL:
						mediaType, err := fetchMediaType(m.ImageURL.URL)
						if err != nil {
							return nil, err
						}
						data, err := urlToBase64(m.ImageURL.URL)
						if err != nil {
							return nil, err
						}
						content[j] = anthropic.NewImageMessageContent(
							anthropic.MessageContentImageSource{
								Type:      "base64",
								MediaType: mediaType,
								Data:      data,
							},
						)
					}
				}
				messages[i] = anthropic.Message{
					Role:    RoleUser,
					Content: content,
				}
			}
		case RoleAssistant:
			messages[i] = anthropic.NewAssistantTextMessage(msg.Content)
		default:
			return nil, fmt.Errorf("unsupported role used for Anthropic: %s", msg.Role)
		}
	}

	return &messages, nil
}
