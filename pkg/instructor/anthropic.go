package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	anthropic "github.com/liushuangls/go-anthropic/v2"
)

type AnthropicClient struct {
	Name string

	client *anthropic.Client
}

var _ Client = &AnthropicClient{}

func NewAnthropicClient(client *anthropic.Client) (*AnthropicClient, error) {
	o := &AnthropicClient{
		Name:   "Anthropic",
		client: client,
	}
	return o, nil
}

func (a *AnthropicClient) Chat(ctx context.Context, request interface{}, mode Mode, schema *Schema) (string, error) {

	req, ok := request.(anthropic.MessagesRequest)
	if !ok {
		return "", fmt.Errorf("invalid request type for %s client", a.Name)
	}

	if !req.Stream {
		return "", errors.New("streaming is not supported by this method; use CreateChatCompletionStream instead")
	}

	switch mode {
	case ModeToolCall:
		return a.completionToolCall(ctx, &req, schema)
	case ModeJSONSchema:
		return a.completionJSONSchema(ctx, &req, schema)
	default:
		return "", fmt.Errorf("mode '%s' is not supported for %s", mode, a.Name)
	}
}

func (a *AnthropicClient) completionToolCall(ctx context.Context, request *anthropic.MessagesRequest, schema *Schema) (string, error) {

	request.Tools = []anthropic.ToolDefinition{}

	for _, function := range schema.Functions {
		t := anthropic.ToolDefinition{
			Name:        function.Name,
			Description: function.Description,
			InputSchema: function.Parameters,
		}
		request.Tools = append(request.Tools, t)
	}

	resp, err := a.client.CreateMessages(ctx, *request)
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

func (a *AnthropicClient) completionJSONSchema(ctx context.Context, request *anthropic.MessagesRequest, schema *Schema) (string, error) {

	system := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself.
`, schema.String)

	if request.System == "" {
		request.System = system
	} else {
		request.System += system
	}

	resp, err := a.client.CreateMessages(ctx, *request)
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
