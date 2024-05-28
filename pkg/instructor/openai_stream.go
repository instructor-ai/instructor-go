package instructor

import (
	"context"
	"errors"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

func (o *OpenAIClient) ChatStream(ctx context.Context, request interface{}, mode Mode, schema *Schema) (<-chan string, error) {

	req, ok := request.(openai.ChatCompletionRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for %s client", o.Name)
	}

	if !req.Stream {
		return nil, errors.New("streaming is not enabled in request type; use CreateChatCompletion for synchronous completion")
	}

	switch mode {
	case ModeToolCall:
		return o.completionToolCallStream(ctx, &req, schema)
	case ModeJSON:
		return o.completionJSONStream(ctx, &req, schema)
	case ModeJSONSchema:
		return o.completionJSONSchemaStream(ctx, &req, schema)
	default:
		return nil, fmt.Errorf("mode '%s' is not supported for %s", mode, o.Name)
	}
}

func (o *OpenAIClient) completionToolCallStream(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (<-chan string, error) {
	request.Tools = createTools(schema)
	return o.createStream(ctx, request)
}

func (o *OpenAIClient) completionJSONStream(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (<-chan string, error) {
	request.Messages = prepend(request.Messages, *createJSONMessageStream(schema))
	// Set JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}
	return o.createStream(ctx, request)
}

func (o *OpenAIClient) completionJSONSchemaStream(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (<-chan string, error) {
	request.Messages = prepend(request.Messages, *createJSONMessageStream(schema))
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

func createJSONMessageStream(schema *Schema) *Message {
	message := fmt.Sprintf(`
Please respond with a JSON array where the elements following JSON schema:

%s

Make sure to return an array with the elements an instance of the JSON, not the schema itself.
`, schema.String)
	return &Message{
		Role:    RoleSystem,
		Content: message,
	}
}

func (o *OpenAIClient) createStream(ctx context.Context, request *openai.ChatCompletionRequest) (<-chan string, error) {
	stream, err := o.Client.CreateChatCompletionStream(ctx, *request)
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
				return
			}
			if err != nil {
				return
			}
			text := response.Choices[0].Delta.Content
			ch <- text
		}
	}()
	return ch, nil
}
