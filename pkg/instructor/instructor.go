package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	anthropic "github.com/liushuangls/go-anthropic/v2"
	openai "github.com/sashabaranov/go-openai"
)

type Instructor struct {
	Client     Client
	Provider   Provider
	Mode       Mode
	MaxRetries int
}

func FromOpenAI(client *openai.Client, opts ...Options) (*Instructor, error) {

	options := mergeOptions(opts...)

	cli, err := NewOpenAIClient(client)
	if err != nil {
		return nil, err
	}

	i := &Instructor{
		Client:     cli,
		Provider:   OpenAI,
		Mode:       *options.Mode,
		MaxRetries: *options.MaxRetries,
	}
	return i, nil
}

func FromAnthropic(client *anthropic.Client, opts ...Options) (*Instructor, error) {

	options := mergeOptions(opts...)

	cli, err := NewAnthropicClient(client)
	if err != nil {
		return nil, err
	}

	i := &Instructor{
		Client:     cli,
		Provider:   OpenAI,
		Mode:       *options.Mode,
		MaxRetries: *options.MaxRetries,
	}
	return i, nil
}

func (i *Instructor) CreateChatCompletion(ctx context.Context, request Request, response any) error {

	t := reflect.TypeOf(response)

	schema, err := NewSchema(t)
	if err != nil {
		return err
	}

	for attempt := 0; attempt < i.MaxRetries; attempt++ {

		text, err := i.Client.CreateChatCompletion(ctx, request, i.Mode, schema)
		if err != nil {
			// no retry on non-marshalling/validation errors
			// return err
			continue
		}

		text = extractJSON(text)

		err = processResponse(text, &response)
		if err != nil {
			// TODO:
			// add more sophisticated retry logic (send back json and parse error for model to fix).
			//
			// Currently, its just recalling with no new information
			// or attempt to fix the error with the last generated JSON
			continue
		}

		return nil
	}

	return errors.New("hit max retry attempts")
}

func processResponse(responseStr string, response *any) error {

	err := json.Unmarshal([]byte(responseStr), response)
	if err != nil {
		return err
	}

	// TODO: if direct unmarshal fails: check common erors like wrapping struct with key name of struct, instead of just the value

	return nil
}

// Removes any prefixes before the JSON (like "Sure, here you go:")
func trimPrefix(jsonStr string) string {
	startObject := strings.IndexByte(jsonStr, '{')
	startArray := strings.IndexByte(jsonStr, '[')

	var start int
	if startObject == -1 && startArray == -1 {
		return jsonStr // No opening brace or bracket found, return the original string
	} else if startObject == -1 {
		start = startArray
	} else if startArray == -1 {
		start = startObject
	} else {
		start = min(startObject, startArray)
	}

	return jsonStr[start:]
}

// Removes any postfixes after the JSON
func trimPostfix(jsonStr string) string {
	endObject := strings.LastIndexByte(jsonStr, '}')
	endArray := strings.LastIndexByte(jsonStr, ']')

	var end int
	if endObject == -1 && endArray == -1 {
		return jsonStr // No closing brace or bracket found, return the original string
	} else if endObject == -1 {
		end = endArray
	} else if endArray == -1 {
		end = endObject
	} else {
		end = max(endObject, endArray)
	}

	return jsonStr[:end+1]
}

// Extracts the JSON by trimming prefixes and postfixes
func extractJSON(jsonStr string) string {
	trimmedPrefix := trimPrefix(jsonStr)
	trimmedJSON := trimPostfix(trimmedPrefix)
	return trimmedJSON
}
