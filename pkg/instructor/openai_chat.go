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
	case ModeStructuredOutputs:
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

	// var s []byte
	// s, _ = json.MarshalIndent(schema, "", "  ")
	// fmt.Println(string(s))

	// oaiSchema := convertToOpenAIJSONSchema(schema.Schema)
	// fmt.Printf(`
	// Type                    %v
	// Description             %v
	// Enum                    %v
	// Properties              %v
	// Required                %v
	// Items                   %v
	// AdditionalProperties    %v
	// `,
	// 	oaiSchema.Type,
	// 	oaiSchema.Description,
	// 	oaiSchema.Enum,
	// 	oaiSchema.Properties,
	// 	oaiSchema.Required,
	// 	oaiSchema.Items,
	// 	oaiSchema.AdditionalProperties,
	// )
	// s, _ = json.MarshalIndent(oaiSchema, "", "  ")
	// fmt.Println(string(s))

	structName := schema.NameFromRef()

	type SchemaWrapper struct {
		Type                 string                  `json:"type"`
		Required             []string                `json:"required"`
		AdditionalProperties bool                    `json:"additionalProperties"`
		Properties           *jsonschema.Definitions `json:"properties"`
		Definitions          *jsonschema.Definitions `json:"$defs"`
	}

	required := []string{structName}
	// // for p := schema.Definitions.; p != nil; p.Next() {
	// for k := range schema.Definitions {
	// 	required = append(required, k)
	// }

	properties := make(jsonschema.Definitions)
	properties[structName] = schema.Definitions[structName]

	schemaWrapper := SchemaWrapper{
		Type:                 "object",
		Required:             required,
		AdditionalProperties: false,
		Properties:           &properties,
		Definitions:          &schema.Schema.Definitions,
	}

	rawSchema, _ := json.Marshal(schemaWrapper)

	request.ResponseFormat = &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
		JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
			Name:        schema.NameFromRef(),
			Description: schema.Description,
			Strict:      true,
			SchemaRaw:   toPtr(rawSchema),
		},
	}

	resp, err := i.Client.CreateChatCompletion(ctx, *request)
	if err != nil {
		return "", nil, err
	}

	text := resp.Choices[0].Message.Content

	// TODO:
	/*
				Get struct contents inside:
				    {
						"MyStructName": {
						    ... // what we want to marshall into struct
						}
		            }
	*/
	resMap := make(map[string]any)
	_ = json.Unmarshal([]byte(text), &resMap)

	cleanedText, _ := json.Marshal(resMap[structName])
	text = string(cleanedText)

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

func convertToOpenAIJSONSchema(schema *jsonschema.Schema) *openaiJSONSchema.Definition {

	oaiSchema := openaiJSONSchema.Definition{}

	// Initialize properties map
	oaiSchema.Properties = make(map[string]openaiJSONSchema.Definition)

	// Convert type; default to object
	if schema.Type != "" {
		oaiSchema.Type = openaiJSONSchema.DataType(schema.Type)
	} else {
		oaiSchema.Type = openaiJSONSchema.Object
	}

	// Convert description
	oaiSchema.Description = schema.Description

	// Convert enum
	if schema.Enum != nil {
		oaiSchema.Enum = make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			oaiSchema.Enum[i] = fmt.Sprintf("%v", v)
		}
	}

	// Convert properties
	if schema.Properties != nil {
		for p := schema.Properties.Oldest(); p != nil; p = p.Next() {
			key, value := p.Key, p.Value
			propertySchema := convertToOpenAIJSONSchema(value)
			oaiSchema.Properties[key] = *propertySchema
		}
	}

	// Convert items
	if schema.Items != nil {
		itemsSchema := convertToOpenAIJSONSchema(schema.Items)
		oaiSchema.Items = itemsSchema
	}

	// Convert additional properties
	if schema.AdditionalProperties != nil {
		additionalPropertiesSchema := convertToOpenAIJSONSchema(schema.AdditionalProperties)
		oaiSchema.AdditionalProperties = additionalPropertiesSchema
	}

	// Convert defintions
	if schema.Definitions != nil {
		for key, value := range schema.Definitions {
			oaiSchema.Required = append(oaiSchema.Required, key)
			fmt.Printf("%+v\n", oaiSchema.Required)

			definitionSchema := convertToOpenAIJSONSchema(value)
			oaiSchema.Properties[key] = *definitionSchema
		}
	}

	if len(oaiSchema.Properties) > 0 {
		oaiSchema.Required = []string{}
		for key := range oaiSchema.Properties {
			oaiSchema.Required = append(oaiSchema.Required, key)
		}
	}

	return &oaiSchema
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
