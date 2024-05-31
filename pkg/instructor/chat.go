package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
)

func chatHandler(i Instructor, ctx context.Context, request interface{}, response any) (interface{}, error) {

	var err error

	t := reflect.TypeOf(response)

	schema, err := NewSchema(t)
	if err != nil {
		return nil, err
	}

	for attempt := 0; attempt < i.MaxRetries(); attempt++ {

		text, resp, err := i.chat(ctx, request, schema)
		if err != nil {
			// no retry on non-marshalling/validation errors
			return nil, err
		}

		text = extractJSON(&text)

		err = json.Unmarshal([]byte(text), &response)
		if err != nil {
			// TODO:
			// add more sophisticated retry logic (send back json and parse error for model to fix).
			//
			// Currently, its just recalling with no new information
			// or attempt to fix the error with the last generated JSON
			continue
		}

		return resp, nil
	}

	return nil, errors.New("hit max retry attempts")
}
