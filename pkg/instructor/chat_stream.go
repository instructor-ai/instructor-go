package instructor

import (
	"context"
	"encoding/json"
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

	ch, err := i.chatStream(ctx, request, schema)
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
