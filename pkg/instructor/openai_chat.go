package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/invopop/jsonschema"
	openai "github.com/sashabaranov/go-openai"
	openaiJSONSchema "github.com/sashabaranov/go-openai/jsonschema"
)

func (i *InstructorOpenAI) CreateChatCompletion(
	ctx context.Context,
	request openai.ChatCompletionRequest,
	responseType any,
) (response openai.ChatCompletionResponse, err error) {

	resp, err := chatHandler(i, ctx, request, responseType)
	if err != nil {
		if resp == nil {
			return openai.ChatCompletionResponse{}, err
		}
		return *nilOpenaiRespWithUsage(resp.(*openai.ChatCompletionResponse)), err
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
	case ModeJSONStrict:
		return i.chatJSONStrict(ctx, &req, schema)
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
		return "", nilOpenaiRespWithUsage(&resp), errors.New("received no tool calls from model, expected at least 1")
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
			return "", nilOpenaiRespWithUsage(&resp), err
		}
		jsonArray[i] = jsonObj
	}

	resultJSON, err := json.Marshal(jsonArray)
	if err != nil {
		return "", nilOpenaiRespWithUsage(&resp), err
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

func (i *InstructorOpenAI) chatJSONStrict(ctx context.Context, request *openai.ChatCompletionRequest, schema *Schema) (string, *openai.ChatCompletionResponse, error) {

	request.Messages = prepend(request.Messages, *createJSONMessage(schema))

	// Set strict JSON mode
	request.ResponseFormat = &openai.ChatCompletionResponseFormat{
		Type:       openai.ChatCompletionResponseFormatTypeJSONObject,
		JSONSchema: convertToChatCompletionResponseFormatJSONSchema(schema, true),
	}

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

func convertSchemaToDefinition(schema *jsonschema.Schema) *openaiJSONSchema.Definition {
	definition := openaiJSONSchema.Definition{
		Type:        openaiJSONSchema.DataType(schema.Type),
		Description: schema.Description,
		Required:    schema.Required,
	}

	if len(schema.Enum) > 0 {
		definition.Enum = make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			definition.Enum[i] = v.(string)
		}
	}

	if schema.Properties != nil {
		definition.Properties = make(map[string]openaiJSONSchema.Definition)
		for el := schema.Properties.Oldest(); el != nil; el = el.Next() {
			k, prop := el.Key, el.Value
			definition.Properties[k] = *convertSchemaToDefinition(prop)
		}
	}

	if schema.Items != nil {
		definition.Items = convertSchemaToDefinition(schema.Items)
	}

	if schema.AdditionalProperties != nil {
		definition.AdditionalProperties = convertSchemaToDefinition(schema.AdditionalProperties)
	}

	return &definition
}

func convertToChatCompletionResponseFormatJSONSchema(schema *Schema, strict bool) *openai.ChatCompletionResponseFormatJSONSchema {
	return &openai.ChatCompletionResponseFormatJSONSchema{
		Name:        schema.Name,
		Description: schema.Description,
		Schema:      *convertSchemaToDefinition(schema.Schema),
		Strict:      strict,
	}
}

func (i *InstructorOpenAI) emptyResponseWithUsageSum(usage *UsageSum) interface{} {
	return &openai.ChatCompletionResponse{
		Usage: openai.Usage{
			PromptTokens:     usage.InputTokens,
			CompletionTokens: usage.OutputTokens,
			TotalTokens:      usage.TotalTokens,
		},
	}
}

func (i *InstructorOpenAI) emptyResponseWithResponseUsage(response interface{}) interface{} {
	resp, ok := response.(*openai.ChatCompletionResponse)
	if !ok || resp == nil {
		return nil
	}

	return &openai.ChatCompletionResponse{
		Usage: resp.Usage,
	}
}

func (i *InstructorOpenAI) addUsageSumToResponse(response interface{}, usage *UsageSum) (interface{}, error) {
	resp, ok := response.(*openai.ChatCompletionResponse)
	if !ok {
		return response, fmt.Errorf("internal type error: expected *openai.ChatCompletionResponse, got %T", response)
	}

	resp.Usage.PromptTokens += usage.InputTokens
	resp.Usage.CompletionTokens += usage.OutputTokens
	resp.Usage.TotalTokens += usage.TotalTokens

	return response, nil
}

func (i *InstructorOpenAI) countUsageFromResponse(response interface{}, usage *UsageSum) *UsageSum {
	resp, ok := response.(*openai.ChatCompletionResponse)
	if !ok {
		return usage
	}

	usage.InputTokens += resp.Usage.PromptTokens
	usage.OutputTokens += resp.Usage.CompletionTokens
	usage.TotalTokens += resp.Usage.TotalTokens

	return usage
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

func nilOpenaiRespWithUsage(resp *openai.ChatCompletionResponse) *openai.ChatCompletionResponse {
	if resp == nil {
		return nil
	}

	return &openai.ChatCompletionResponse{
		Usage: resp.Usage,
	}
}
