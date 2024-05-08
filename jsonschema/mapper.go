package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/distantmagic/paddler/llamacpp"
	santoshiSchema "github.com/santhosh-tekuri/jsonschema/v5"
)

type JsonSchemaMapper struct {
	LlamaCppClient *llamacpp.LlamaCppClient
}

func (self *JsonSchemaMapper) MapToSchema(
	responseChannel chan JsonSchemaMapperResult,
	jsonSchema any,
	userInput string,
) {
	defer close(responseChannel)

	marshaledSchema, err := json.Marshal(jsonSchema)

	if err != nil {
		responseChannel <- JsonSchemaMapperResult{
			Error: err,
		}

		return
	}

	jsonSchemaCompiler := santoshiSchema.NewCompiler()

	err = jsonSchemaCompiler.AddResource(
		"schema.json",
		bytes.NewReader(marshaledSchema),
		// strings.NewReader(string(marshaledSchema)),
	)

	if err != nil {
		responseChannel <- JsonSchemaMapperResult{
			Error: err,
		}

		return
	}

	schema, err := jsonSchemaCompiler.Compile("schema.json")

	if err != nil {
		responseChannel <- JsonSchemaMapperResult{
			Error: err,
		}

		return
	}

	llamaCppCompletionResponseChannel := make(chan llamacpp.LlamaCppCompletionToken)

	go self.LlamaCppClient.GenerateCompletion(
		llamaCppCompletionResponseChannel,
		llamacpp.LlamaCppCompletionRequest{
			JsonSchema: jsonSchema,
			NPredict: 100,
			Prompt: fmt.Sprintf(
				`User will provide the phrase. Respond with JSON matching the
				schema. Fill the schema with the infromation provided in the
				user phrase.

				---
				JSON schema:
				%s
				---

				---
				User phrase:
				%s
				---`,
				marshaledSchema,
				userInput,
			),
			Stream: true,
		},
	)

	acc := ""

	for token := range llamaCppCompletionResponseChannel {
		if token.Error != nil {
			responseChannel <- JsonSchemaMapperResult{
				Error: token.Error,
			}

			return
		}

		acc += token.Content
	}

	var unmarshaledLlamaResponse any

	err = json.Unmarshal([]byte(acc), &unmarshaledLlamaResponse)

	if err != nil {
		responseChannel <- JsonSchemaMapperResult{
			Error: err,
		}

		return
	}

	err = schema.Validate(unmarshaledLlamaResponse)

	if err != nil {
		responseChannel <- JsonSchemaMapperResult{
			Error: err,
		}

		return
	}

	responseChannel <- JsonSchemaMapperResult{
		Result: unmarshaledLlamaResponse,
	}
}
