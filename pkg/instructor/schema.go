package instructor

import (
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
)

type Schema[T any] struct {
	*jsonschema.Schema
	String string

	Functions []FunctionDefinition
}

type Function struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Parameters  *jsonschema.Schema `json:"parameters"`
}

func NewSchema[T any]() (*Schema[T], error) {

	t := new(T)

	tType := reflect.TypeOf(t)

	schema := jsonschema.ReflectFromType(tType)

	str, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	funcs := ToFunctionSchema(tType, schema)

	s := &Schema[T]{
		Schema: schema,
		String: string(str),

		Functions: funcs,
	}

	return s, nil
}

func ToFunctionSchema(tType reflect.Type, tSchema *jsonschema.Schema) []FunctionDefinition {

	fds := []FunctionDefinition{}

	for name, def := range tSchema.Definitions {

		parameters := &jsonschema.Schema{
			Type:       "object",
			Properties: def.Properties,
			Required:   def.Required,
		}

		fd := FunctionDefinition{
			Name:        name,
			Description: def.Description,
			Parameters:  parameters,
		}

		fds = append(fds, fd)
	}

	return fds
}
