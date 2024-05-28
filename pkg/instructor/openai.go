package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	Name Provider

	*openai.Client
}

var _ Client = &OpenAIClient{}

func NewOpenAIClient(client *openai.Client) (*OpenAIClient, error) {
	o := &OpenAIClient{
		Name:   "OpenAI",
		Client: client,
	}
	return o, nil
}

func (o *OpenAIClient) Chat(ctx context.Context, request interface{}, mode Mode, schema *Schema) (string, error) {

	req, ok := request.(openai.ChatCompletionRequest)
	if !ok {
		return "", fmt.Errorf("invalid request type for %s client", o.Name)
	}

	if req.Stream {
		return "", errors.New("streaming is not supported by this method; use CreateChatCompletionStream instead")
	}

	switch mode {
	case ModeToolCall:
		return o.completionToolCall(ctx, &req, schema)
	case ModeJSON:
		return o.completionJSON(ctx, &req, schema)
	case ModeJSONSchema:
		return o.completionJSONSchema(ctx, &req, schema)
	default:
		return "", fmt.Errorf("mode '%s' is not supported for %s", mode, o.Name)
	}
}

func (o *OpenAIClient) completionToolCall(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (string, error) {

	request.Tools = createTools(schema)

	resp, err := o.Client.CreateChatCompletion(ctx, *request)
	if err != nil {
		return "", err
	}

	var toolCalls []openai.ToolCall
	for _, choice := range resp.Choices {
		toolCalls = choice.Message.ToolCalls

		if len(toolCalls) >= 1 {
			break
		}
	}

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

func (o *OpenAIClient) completionJSON(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (string, error) {

	request.Messages = prepend(request.Messages, *createJSONMessage(schema))

	// Set JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}

	resp, err := o.Client.CreateChatCompletion(ctx, *request)
	if err != nil {
		return "", err
	}

	text := resp.Choices[0].Message.Content

	return text, nil
}

func (o *OpenAIClient) completionJSONSchema(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (string, error) {

	request.Messages = prepend(request.Messages, *createJSONMessage(schema))

	resp, err := o.Client.CreateChatCompletion(ctx, *request)
	if err != nil {
		return "", err
	}

	text := resp.Choices[0].Message.Content

	return text, nil
}

func createJSONMessage(schema *Schema) *Message {
	message := fmt.Sprintf(`
Please respond with JSON in the following JSON schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, schema.String)
	return &Message{
		Role:    RoleSystem,
		Content: message,
	}
}
