package instructor

import (
	"context"
	"errors"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

func (i *InstructorOpenAI) CreateChatCompletionStream(
	ctx context.Context,
	request openai.ChatCompletionRequest,
	responseType any,
) (stream <-chan any, err error) {

	stream, err = chatStreamHandler(i, ctx, request, responseType)
	if err != nil {
		return nil, err
	}

	return stream, err
}

func (i *InstructorOpenAI) chatStream(ctx context.Context, request interface{}, schema *Schema) (<-chan string, error) {

	req, ok := request.(openai.ChatCompletionRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type for %s client", i.Provider())
	}

	if !req.Stream {
		return nil, errors.New("streaming is not enabled in request type; use CreateChatCompletion for synchronous completion")
	}

	switch i.Mode() {
	case ModeToolCall:
		return i.chatToolCallStream(ctx, &req, schema)
	case ModeJSON:
		return i.chatJSONStream(ctx, &req, schema)
	case ModeJSONSchema:
		return i.chatJSONSchemaStream(ctx, &req, schema)
	default:
		return nil, fmt.Errorf("mode '%s' is not supported for %s", i.Mode(), i.Provider())
	}
}

func (i *InstructorOpenAI) chatToolCallStream(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (<-chan string, error) {
	request.Tools = createTools(schema)
	return i.createStream(ctx, request)
}

func (i *InstructorOpenAI) chatJSONStream(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (<-chan string, error) {
	request.Messages = prepend(request.Messages, *createJSONMessageStream(schema))
	// Set JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}
	return i.createStream(ctx, request)
}

func (i *InstructorOpenAI) chatJSONSchemaStream(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (<-chan string, error) {
	request.Messages = prepend(request.Messages, *createJSONMessageStream(schema))
	return i.createStream(ctx, request)
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

func createJSONMessageStream(schema *Schema) *openai.ChatCompletionMessage {
	message := fmt.Sprintf(`
Please respond with a JSON array where the elements following JSON schema:

%s

Make sure to return an array with the elements an instance of the JSON, not the schema itself.
`, schema.String)

	msg := &openai.ChatCompletionMessage{
		Role:    RoleSystem,
		Content: message,
	}

	return msg
}

func (i *InstructorOpenAI) createStream(ctx context.Context, request *openai.ChatCompletionRequest) (<-chan string, error) {
	stream, err := i.Client.CreateChatCompletionStream(ctx, *request)
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
