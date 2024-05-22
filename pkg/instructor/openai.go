package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	Name string

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

func (o *OpenAIClient) CreateChatCompletion(ctx context.Context, request Request, mode Mode, schema *Schema) (string, error) {
	if request.Stream {
		return "", errors.New("streaming is not supported by this method; use CreateChatCompletionStream instead")
	}

	switch mode {
	case ModeToolCall:
		return o.completionToolCall(ctx, request, schema)
	case ModeJSON:
		return o.completionJSON(ctx, request, schema)
	case ModeJSONSchema:
		return o.completionJSONSchema(ctx, request, schema)
	default:
		return "", fmt.Errorf("mode '%s' is not supported for %s", mode, o.Name)
	}
}

func (o *OpenAIClient) completionToolCall(ctx context.Context, request Request, schema *Schema) (string, error) {

	request.Tools = createTools(schema)

	resp, err := o.Client.CreateChatCompletion(ctx, request)
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

func (o *OpenAIClient) completionJSON(ctx context.Context, request Request, schema *Schema) (string, error) {

	request.Messages = prepend(request.Messages, *createJSONMessage(schema))

	// Set JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}

	resp, err := o.Client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	text := resp.Choices[0].Message.Content

	return text, nil
}

func (o *OpenAIClient) completionJSONSchema(ctx context.Context, request Request, schema *Schema) (string, error) {

	request.Messages = prepend(request.Messages, *createJSONMessage(schema))

	resp, err := o.Client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	text := resp.Choices[0].Message.Content

	return text, nil
}

func (o *OpenAIClient) CreateChatCompletionStream(ctx context.Context, request Request, mode Mode, schema *Schema) (<-chan string, error) {
	switch mode {
	case ModeToolCall:
		return o.completionToolCallStream(ctx, request, schema)
	case ModeJSON:
		return o.completionJSONStream(ctx, request, schema)
	case ModeJSONSchema:
		return o.completionJSONSchemaStream(ctx, request, schema)
	default:
		return nil, fmt.Errorf("mode '%s' is not supported for %s", mode, o.Name)
	}
}

func (o *OpenAIClient) completionToolCallStream(ctx context.Context, request Request, schema *Schema) (<-chan string, error) {
	request.Tools = createTools(schema)
	return o.createStream(ctx, request)
}

func (o *OpenAIClient) completionJSONStream(ctx context.Context, request Request, schema *Schema) (<-chan string, error) {
	request.Messages = prepend(request.Messages, *createJSONMessage(schema))
	// Set JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}
	return o.createStream(ctx, request)
}

func (o *OpenAIClient) completionJSONSchemaStream(ctx context.Context, request Request, schema *Schema) (<-chan string, error) {
	request.Messages = prepend(request.Messages, *createJSONMessage(schema))
	return o.createStream(ctx, request)
}

func createTools(schema *Schema) []openai.Tool {
	tools := make([]openai.Tool, 0, len(schema.Functions))
	for _, function := range schema.Functions {
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
	return tools
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

func (o *OpenAIClient) createStream(ctx context.Context, request Request) (<-chan string, error) {
	stream, err := o.Client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, err
	}

	ch := make(chan string)

	go func() {
		defer stream.Close()
		defer close(ch)
		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
                println("closing channel")
				return
			}
			if err != nil {
                println("channel errored")
				return
			}
			text := response.Choices[0].Delta.Content
			ch <- text
		}
	}()
	return ch, nil
}
