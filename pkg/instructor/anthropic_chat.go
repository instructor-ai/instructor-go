package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	anthropic "github.com/liushuangls/go-anthropic/v2"
)

func (i *InstructorAnthropic) CreateMessages(ctx context.Context, request anthropic.MessagesRequest, responseType any) (response anthropic.MessagesResponse, err error) {

	resp, err := chatHandler(i, ctx, request, responseType)
	if err != nil {
		return anthropic.MessagesResponse{}, err
	}

	response = *(resp.(*anthropic.MessagesResponse))

	return response, nil
}

func (i *InstructorAnthropic) chat(ctx context.Context, request interface{}, schema *Schema) (string, interface{}, error) {

	req, ok := request.(anthropic.MessagesRequest)
	if !ok {
		return "", nil, fmt.Errorf("invalid request type for %s client", i.Provider())
	}

	if req.Stream {
		return "", nil, errors.New("streaming is not supported by this method; use CreateChatCompletionStream instead")
	}

	switch i.Mode() {
	case ModeToolCall:
		return i.completionToolCall(ctx, &req, schema)
	case ModeJSONSchema:
		return i.completionJSONSchema(ctx, &req, schema)
	default:
		return "", nil, fmt.Errorf("mode '%s' is not supported for %s", i.Mode(), i.Provider())
	}
}

func (i *InstructorAnthropic) completionToolCall(ctx context.Context, request *anthropic.MessagesRequest, schema *Schema) (string, *anthropic.MessagesResponse, error) {

	request.Tools = []anthropic.ToolDefinition{}

	for _, function := range schema.Functions {
		t := anthropic.ToolDefinition{
			Name:        function.Name,
			Description: function.Description,
			InputSchema: function.Parameters,
		}
		request.Tools = append(request.Tools, t)
	}

	resp, err := i.Client.CreateMessages(ctx, *request)
	if err != nil {
		return "", nil, err
	}

	for _, c := range resp.Content {
		if c.Type != anthropic.MessagesContentTypeToolUse {
			// Skip non tool responses
			continue
		}

		toolInput, err := json.Marshal(c.Input)
		if err != nil {
			return "", nil, err
		}
		// TODO: handle more than 1 tool use
		return string(toolInput), &resp, nil
	}

	return "", nil, errors.New("more than 1 tool response at a time is not implemented")

}

func (i *InstructorAnthropic) completionJSONSchema(ctx context.Context, request *anthropic.MessagesRequest, schema *Schema) (string, *anthropic.MessagesResponse, error) {

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

	resp, err := i.Client.CreateMessages(ctx, *request)
	if err != nil {
		return "", nil, err
	}

	text := resp.Content[0].Text

	return *text, &resp, nil
}
