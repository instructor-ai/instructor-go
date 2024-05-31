package instructor

import (
	"context"
	"fmt"

	cohere "github.com/cohere-ai/cohere-go/v2"
	option "github.com/cohere-ai/cohere-go/v2/option"
)

func (i *InstructorCohere) Chat(
	ctx context.Context,
	request *cohere.ChatRequest,
	response any,
	opts ...option.RequestOption,
) (*cohere.NonStreamedChatResponse, error) {

	resp, err := chatHandler(i, ctx, request, response)
	if err != nil {
		return nil, err
	}

	return resp.(*cohere.NonStreamedChatResponse), nil
}

func (i *InstructorCohere) chat(ctx context.Context, request interface{}, schema *Schema) (string, interface{}, error) {

	req, ok := request.(*cohere.ChatRequest)
	if !ok {
		return "", nil, fmt.Errorf("invalid request type for %s client", i.Provider())
	}

	switch i.Mode() {
	case ModeToolCall:
		return i.chatToolCall(ctx, req, schema)
	case ModeJSON:
		return i.chatJSON(ctx, req, schema)
	default:
		return "", nil, fmt.Errorf("mode '%s' is not supported for %s", i.Mode(), i.Provider())
	}
}

func (i *InstructorCohere) chatToolCall(ctx context.Context, request *cohere.ChatRequest, schema *Schema) (string, *cohere.NonStreamedChatResponse, error) {

	request.Tools = []*cohere.Tool{createCohereTools(schema)}

	resp, err := i.Client.Chat(ctx, request)
	if err != nil {
		return "", nil, err
	}

	_ = resp

	// TODO: implement

	panic("tool call not implemented Cohere")

}

func (i *InstructorCohere) chatJSON(ctx context.Context, request *cohere.ChatRequest, schema *Schema) (string, *cohere.NonStreamedChatResponse, error) {

	i.addOrConcatJSONSystemPrompt(request, schema)

	resp, err := i.Client.Chat(ctx, request)
	if err != nil {
		return "", nil, err
	}

	return resp.Text, resp, nil
}

func (i *InstructorCohere) addOrConcatJSONSystemPrompt(request *cohere.ChatRequest, schema *Schema) {

	schemaPrompt := fmt.Sprintf("```json!Please respond with JSON in the following JSON schema - make sure to return an instance of the JSON, not the schema itself: %s ", schema.String)

	if request.Preamble == nil {
		request.Preamble = &schemaPrompt
	} else {
		request.Preamble = toPtr(*request.Preamble + "\n" + schemaPrompt)
	}
}

func createCohereTools(schema *Schema) *cohere.Tool {

	tool := &cohere.Tool{
		Name:                 "functions",
		Description:          schema.Schema.Description,
		ParameterDefinitions: make(map[string]*cohere.ToolParameterDefinitionsValue),
	}

	for _, function := range schema.Functions {
		parameterDefinition := &cohere.ToolParameterDefinitionsValue{
			Description: toPtr(function.Description),
			Type:        function.Parameters.Type,
		}
		tool.ParameterDefinitions[function.Name] = parameterDefinition
	}

	return tool
}
