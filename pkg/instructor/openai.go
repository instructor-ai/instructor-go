package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient[T any] struct {
	Name string

	client *openai.Client
	schema *Schema[T]
	mode   Mode
}

var _ Client[any] = &OpenAIClient[any]{}

func NewOpenAIClient[T any](client *openai.Client, schema *Schema[T], mode Mode) (*OpenAIClient[T], error) {
	o := &OpenAIClient[T]{
		Name:   "OpenAI",
		client: client,
		schema: schema,
		mode:   mode,
	}
	return o, nil
}

func (o *OpenAIClient[any]) CreateChatCompletion(ctx context.Context, request Request) (string, error) {
	return o.completionModeHandler(ctx, request)
}

func (o *OpenAIClient[any]) completionModeHandler(ctx context.Context, request Request) (string, error) {
	switch o.mode {
	case ModeToolCall:
		return o.completionToolCall(ctx, request)
	case ModeJSON:
		return o.completionJSON(ctx, request)
	case ModeJSONSchema:
		return o.completionJSONSchema(ctx, request)
	default:
		return "", fmt.Errorf("mode '%s' is not supported for %s", o.mode, o.Name)
	}
}

func (o *OpenAIClient[any]) completionToolCall(ctx context.Context, request Request) (string, error) {

	tools := []openai.Tool{}

	for _, function := range o.schema.Functions {
		f := openai.FunctionDefinition{
			Name:        function.Name,
			Description: function.Description,
			Parameters:  function.Parameters,
		}
		t := openai.Tool{
			Type:     "function",
			Function: &f,
		}
		tools = append(tools, t)
	}

	request.Tools = tools

	resp, err := o.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	toolCalls := resp.Choices[0].Message.ToolCalls
	numTools := len(toolCalls)

	if numTools < 1 {
		return "", errors.New("recieved no tool calls from model, expected at least 1")
	}

	if numTools == 1 {
		return toolCalls[0].Function.Arguments, nil
	}

	// numTools >= 1

	jsonArray := make([]map[string]interface{}, len(toolCalls))

	for i, toolCall := range toolCalls {
		var jsonObj map[string]interface{}
		err = json.Unmarshal([]byte(toolCall.Function.Arguments), &jsonObj)
		if err != nil {
			return "", err
		}
		jsonArray[i] = jsonObj
	}

	resultJSON, err := json.Marshal(jsonArray)
	if err != nil {
		return "", err
	}

	return string(resultJSON), nil
}

func (o *OpenAIClient[any]) completionJSON(ctx context.Context, request Request) (string, error) {
	message := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, o.schema.String)

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

func (o *OpenAIClient[any]) completionJSONSchema(ctx context.Context, request Request) (string, error) {

	message := fmt.Sprintf(`
Please responsd with json in the following json_schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, o.schema.String)

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
