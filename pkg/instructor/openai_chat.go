package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func (i *InstructorOpenAI) CreateChatCompletion(
	ctx context.Context,
	request openai.ChatCompletionRequest,
	responseType any,
) (response openai.ChatCompletionResponse, err error) {

	resp, err := chatHandler(i, ctx, request, responseType)
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}

	response = *(resp.(*openai.ChatCompletionResponse))

	return response, nil
}

func (i *InstructorOpenAI) chat(ctx context.Context, request interface{}, schema *Schema) (string, interface{}, error) {

	req, ok := request.(openai.ChatCompletionRequest)
	if !ok {
		return "", nil, fmt.Errorf("invalid request type for %s client", i.Provider())
	}

	if req.Stream {
		return "", nil, errors.New("streaming is not supported by this method; use CreateChatCompletionStream instead")
	}

	switch i.Mode() {
	case ModeToolCall:
		return i.chatToolCall(ctx, &req, schema)
	case ModeJSON:
		return i.chatJSON(ctx, &req, schema)
	case ModeJSONSchema:
		return i.chatJSONSchema(ctx, &req, schema)
	default:
		return "", nil, fmt.Errorf("mode '%s' is not supported for %s", i.Mode(), i.Provider())
	}
}

func (i *InstructorOpenAI) chatToolCall(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (string, *openai.ChatCompletionResponse, error) {

	request.Tools = createOpenAITools(schema)

	resp, err := i.Client.CreateChatCompletion(ctx, *request)
	if err != nil {
		return "", nil, err
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
		return "", nil, errors.New("recieved no tool calls from model, expected at least 1")
	}

	if numTools == 1 {
		return toolCalls[0].Function.Arguments, &resp, nil
	}

	// numTools >= 1

	jsonArray := make([]map[string]interface{}, len(toolCalls))

	for i, toolCall := range toolCalls {
		var jsonObj map[string]interface{}
		err = json.Unmarshal([]byte(toolCall.Function.Arguments), &jsonObj)
		if err != nil {
			return "", nil, err
		}
		jsonArray[i] = jsonObj
	}

	resultJSON, err := json.Marshal(jsonArray)
	if err != nil {
		return "", nil, err
	}

	return string(resultJSON), &resp, nil
}

func (i *InstructorOpenAI) chatJSON(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (string, *openai.ChatCompletionResponse, error) {

	request.Messages = prepend(request.Messages, *createJSONMessage(schema))

	// Set JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}

	resp, err := i.Client.CreateChatCompletion(ctx, *request)
	if err != nil {
		return "", nil, err
	}

	text := resp.Choices[0].Message.Content

	return text, &resp, nil
}

func (i *InstructorOpenAI) chatJSONSchema(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (string, *openai.ChatCompletionResponse, error) {

	request.Messages = prepend(request.Messages, *createJSONMessage(schema))

	resp, err := i.Client.CreateChatCompletion(ctx, *request)
	if err != nil {
		return "", nil, err
	}

	text := resp.Choices[0].Message.Content

	return text, &resp, nil
}

func createJSONMessage(schema *Schema) *openai.ChatCompletionMessage {
	message := fmt.Sprintf(`
Please respond with JSON in the following JSON schema:

%s

Make sure to return an instance of the JSON, not the schema itself
`, schema.String)

	msg := &openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	}

	return msg
}
