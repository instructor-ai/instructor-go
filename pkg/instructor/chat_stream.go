package instructor

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type StreamWrapper[T any] struct {
	Items []T `json:"items"`
}

const WRAPPER_END = `"items": [`

func chatStreamHandler(i Instructor, ctx context.Context, request interface{}, response any) (<-chan interface{}, error) {

	validate = validator.New()

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

	ch, err := i.chatStream(ctx, request, schema)
	if err != nil {
		return nil, err
	}

	parsedChan := parseStream(ctx, ch, responseType)

	return parsedChan, nil
}

func parseStream(ctx context.Context, ch <-chan string, responseType reflect.Type) <-chan interface{} {

	parsedChan := make(chan any)

	go func() {
		defer close(parsedChan)

		buffer := new(strings.Builder)
		inArray := false

		for {
			select {
			case <-ctx.Done():
				return
			case text, ok := <-ch:
				if !ok {
					// Stream closed
					processRemainingBuffer(buffer, parsedChan, responseType)
					return
				}

				buffer.WriteString(text)

				// Eat all input until elements stream starts
				if !inArray {
					inArray = startArray(buffer)
				}

				processBuffer(buffer, parsedChan, responseType)
			}
		}
	}()

	return parsedChan
}

func startArray(buffer *strings.Builder) bool {

	data := buffer.String()

	idx := strings.Index(data, WRAPPER_END)
	if idx == -1 {
		return false
	}

	trimmed := strings.TrimSpace(data[idx+len(WRAPPER_END):])
	buffer.Reset()
	buffer.WriteString(trimmed)

	return true
}

func processBuffer(buffer *strings.Builder, parsedChan chan<- interface{}, responseType reflect.Type) {

	data := buffer.String()

	data, remaining := getFirstFullJSONElement(&data)

	decoder := json.NewDecoder(strings.NewReader(data))

	for decoder.More() {
		instance := reflect.New(responseType).Interface()
		err := decoder.Decode(instance)
		if err != nil {
			break
		}

		// Validate the instance
		err = validate.Struct(instance)
		if err != nil {
			break
		}

		parsedChan <- instance

		buffer.Reset()
		buffer.WriteString(remaining)
	}
}

func processRemainingBuffer(buffer *strings.Builder, parsedChan chan<- interface{}, responseType reflect.Type) {

	data := buffer.String()

	data = extractJSON(&data)

	if idx := strings.LastIndex(data, "]"); idx != -1 {
		data = data[:idx]
	}

	processBuffer(buffer, parsedChan, responseType)

}
