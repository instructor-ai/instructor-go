package instructor

import (
	"context"
	"encoding/json"
	"errors"
	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/liushuangls/go-anthropic/v2"
	"github.com/sashabaranov/go-openai"
	"reflect"

	"github.com/go-playground/validator/v10"
)

type UsageSum struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

func chatHandler(i Instructor, ctx context.Context, request interface{}, response any) (interface{}, error) {

	var err error

	t := reflect.TypeOf(response)

	schema, err := NewSchema(t)
	if err != nil {
		return nil, err
	}

	// keep a running total of usage
	usage := &UsageSum{}

	for attempt := 0; attempt < i.MaxRetries(); attempt++ {

		text, resp, err := i.chat(ctx, request, schema)
		if err != nil {
			// no retry on non-marshalling/validation errors
			return nilChatRespWithUsage(resp), err
		}

		text = extractJSON(&text)

		err = json.Unmarshal([]byte(text), &response)
		if err != nil {
			// TODO:
			// add more sophisticated retry logic (send back json and parse error for model to fix).
			//
			// Currently, its just recalling with no new information
			// or attempt to fix the error with the last generated JSON

			countUsageFromResp(resp, usage)
			continue
		}

		if i.Validate() {
			validate = validator.New()
			// Validate the response structure against the defined model using the validator
			err = validate.Struct(response)

			if err != nil {
				// TODO:
				// add more sophisticated retry logic (send back validator error and parse error for model to fix).

				countUsageFromResp(resp, usage)
				continue
			}
		}

		return addUsageSumToResp(resp, usage), nil
	}

	return i.EmptyResponseWithUsage(usage), errors.New("hit max retry attempts")
}

func nilChatRespWithUsage(response interface{}) interface{} {
	switch resp := response.(type) {
	case nil:
		return nil
	case *openai.ChatCompletionResponse:
		return &openai.ChatCompletionResponse{
			Usage: resp.Usage,
		}
	case *anthropic.MessagesResponse:
		return &anthropic.MessagesResponse{
			Usage: resp.Usage,
		}
	case *cohere.NonStreamedChatResponse:
		return &cohere.NonStreamedChatResponse{
			Meta: resp.Meta,
		}
	default:
		return nil
	}
}

func addUsageSumToResp(response interface{}, usage *UsageSum) interface{} {
	switch resp := response.(type) {
	case *openai.ChatCompletionResponse:
		resp.Usage.PromptTokens += usage.InputTokens
		resp.Usage.CompletionTokens += usage.OutputTokens
		resp.Usage.TotalTokens += usage.TotalTokens
	case *anthropic.MessagesResponse:
		resp.Usage.InputTokens += usage.InputTokens
		resp.Usage.OutputTokens += usage.OutputTokens
	case *cohere.NonStreamedChatResponse:
		*resp.Meta.Tokens.InputTokens += float64(usage.InputTokens)
		*resp.Meta.Tokens.OutputTokens += float64(usage.OutputTokens)
	}
	return response
}

func countUsageFromResp(response interface{}, usage *UsageSum) {
	switch resp := response.(type) {
	case *openai.ChatCompletionResponse:
		usage.InputTokens += resp.Usage.PromptTokens
		usage.OutputTokens += resp.Usage.CompletionTokens
		usage.TotalTokens += resp.Usage.TotalTokens
	case *anthropic.MessagesResponse:
		usage.InputTokens += resp.Usage.InputTokens
		usage.OutputTokens += resp.Usage.OutputTokens
	case *cohere.NonStreamedChatResponse:
		usage.InputTokens += int(*resp.Meta.Tokens.InputTokens)
		usage.OutputTokens += int(*resp.Meta.Tokens.OutputTokens)
	}
}
