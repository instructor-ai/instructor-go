package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	cohere "github.com/cohere-ai/cohere-go/v2/client"
	anthropic "github.com/liushuangls/go-anthropic/v2"
	openai "github.com/sashabaranov/go-openai"
)

type Client interface {
	Chat(
		ctx context.Context,
		request interface{},
		mode Mode,
		schema *Schema,
	) (string, error)

	ChatStream(
		ctx context.Context,
		request interface{},
		mode Mode,
		schema *Schema,
	) (<-chan string, error)
}

type Instructor struct {
	Provider   Provider
	Mode       Mode
	MaxRetries int

	Client
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

func FromCohere(client *cohere.Client, opts ...Options) (*Instructor, error) {
	panic("not implemented")
}

func (i *Instructor) Chat(ctx context.Context, request interface{}, response any) error {

	t := reflect.TypeOf(response)

	schema, err := NewSchema(t)
	if err != nil {
		return err
	}

	for attempt := 0; attempt < i.MaxRetries; attempt++ {

		text, err := i.Client.Chat(ctx, request, i.Mode, schema)
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

func (i *Instructor) CreateChatCompletion(ctx context.Context, request interface{}, response any) error {
	return i.Chat(ctx, request, response)
}

func (i *Instructor) ChatStream(ctx context.Context, request interface{}, response any) (chan any, error) {

	const WRAPPER_END = `"items": [`

	type StreamWrapper[T any] struct {
		Items []T `json:"items"`
	}

	responseType := reflect.TypeOf(response)

	streamWrapperType := reflect.StructOf([]reflect.StructField{
		{
			Name:      "Items",
			Type:      reflect.SliceOf(responseType),
			Tag:       `json:"items"`,
			Anonymous: false,
		},
	})

	schema, err := NewSchema(streamWrapperType)
	if err != nil {
		return nil, err
	}

	ch, err := i.Client.ChatStream(ctx, &request, i.Mode, schema)
	if err != nil {
		return nil, err
	}

	parsedChan := make(chan any) // Buffered channel for parsed objects

	go func() {
		defer close(parsedChan)
		var buffer strings.Builder
		inArray := false

		for {
			select {
			case <-ctx.Done():
				return
			case text, ok := <-ch:
				if !ok {
					// Steeam closed

					// Get last element out of stream wrapper

					data := buffer.String()

					if idx := strings.LastIndex(data, "]"); idx != -1 {
						data = data[:idx] + data[idx+1:]
					}

					// Process the remaining data in the buffer
					decoder := json.NewDecoder(strings.NewReader(data))
					for decoder.More() {
						instance := reflect.New(responseType).Interface()
						err := decoder.Decode(instance)
						if err != nil {
							break
						}
						parsedChan <- instance
					}
					return
				}
				buffer.WriteString(text)

				// eat all input until elements stream starts
				if !inArray {
					idx := strings.Index(buffer.String(), WRAPPER_END)
					if idx == -1 {
						continue
					}

					inArray = true
					bufferStr := buffer.String()
					trimmed := strings.TrimSpace(bufferStr[idx+len(WRAPPER_END):])
					buffer.Reset()
					buffer.WriteString(trimmed)
				}

				data := buffer.String()
				decoder := json.NewDecoder(strings.NewReader(data))

				for decoder.More() {
					instance := reflect.New(responseType).Interface()
					err := decoder.Decode(instance)
					if err != nil {
						break
					}
					parsedChan <- instance

					buffer.Reset()
					buffer.WriteString(data[len(data):])
				}
			}
		}
	}()

	return parsedChan, nil
}

func (i *Instructor) CreateChatCompletionStream(ctx context.Context, request interface{}, response any) (chan any, error) {
	return i.ChatStream(ctx, request, response)
}

func processResponse(responseStr string, response *any) error {
	err := json.Unmarshal([]byte(responseStr), response)
	if err != nil {
		return err
	}

	// TODO: if direct unmarshal fails: check common errors like wrapping struct with key name of struct, instead of just the value

	return nil
}
