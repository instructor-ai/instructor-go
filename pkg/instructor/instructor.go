package instructor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"

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

func CreateChatCompletionStream[T any](i *Instructor, ctx context.Context, request Request, ch chan T) error {
	go func() {
		defer close(ch)

		t := reflect.TypeOf(new(T)).Elem()

		schema, err := NewSchema(t)
		if err != nil {
			ch <- *new(T) // send a zero value of type T to signal error
			return
		}

		request.Stream = true

		streamCh, err := i.Client.CreateChatCompletionStream(ctx, request, i.Mode, schema)
		if err != nil {
			ch <- *new(T) // send a zero value of type T to signal error
			return
		}

		var buffer bytes.Buffer

		for {
			select {
			case <-ctx.Done():
				return
			default:
				text, ok := <-streamCh
				if !ok {
					return
				}

				println(text)
				text = extractJSON(text)

				buffer.WriteString(text)

				for {
					var chunk T
					err = json.Unmarshal(buffer.Bytes(), &chunk)
					if err == nil {
						ch <- chunk
						buffer.Reset()
						break
					}

					if err != io.EOF {
						break
					}
				}
			}
		}
	}()

	return nil
}

func processResponse(responseStr string, response *any) error {

	err := json.Unmarshal([]byte(responseStr), response)
	if err != nil {
		return err
	}

	// TODO: if direct unmarshal fails: check common erors like wrapping struct with key name of struct, instead of just the value

	return nil
}
