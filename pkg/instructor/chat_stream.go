package instructor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func chatStreamHandler(i Instructor, ctx context.Context, request interface{}, response any) (<-chan any, error) {

	type StreamWrapper[T any] struct {
		Items []T `json:"items"`
	}

	const WRAPPER_END = `"items": [`

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

	ch, err := i.chatStream(ctx, &request, schema)
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
					for len(data) > 0 {
						parsedObject, remaining, err := parseJSONObjectStream(data, responseType)
						if err != nil {
							break
						}
						parsedChan <- parsedObject
						data = remaining
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
				parsedObject, remaining, err := parseJSONObjectStream(data, responseType)
				if err != nil {
					break
				}
				parsedChan <- parsedObject
				buffer.Reset()
				buffer.WriteString(remaining)
			}
		}
	}()

	return parsedChan, nil
}
func parseJSONObjectStream(jsonStr string, responseType reflect.Type) (interface{}, string, error) {
	decoder := json.NewDecoder(strings.NewReader(jsonStr))

	// Skip the initial '{'
	if _, err := consumeToken(decoder); err != nil {
		return nil, "", err
	}

	instance := reflect.New(responseType).Interface()
	if err := decoder.Decode(instance); err != nil {
		return nil, "", err
	}

	remaining, err := consumeRemaining(decoder)
	if err != nil {
		return nil, "", err
	}

	return instance, remaining, nil
}

func consumeToken(decoder *json.Decoder) (json.Token, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	return token, nil
}

func consumeRemaining(decoder *json.Decoder) (string, error) {
	token, err := decoder.Token()
	if err != nil {
		return "", err
	}

	switch t := token.(type) {
	case json.Delim:
		return string(t), nil
	case string:
		return t, nil
	default:
		// Handle other token types as needed
		return "", fmt.Errorf("unexpected token type: %T", t)
	}
}
