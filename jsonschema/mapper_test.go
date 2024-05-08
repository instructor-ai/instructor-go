package instruct

import (
	"net/http"
	"testing"

	"github.com/distantmagic/paddler/llamacpp"
	"github.com/distantmagic/paddler/netcfg"
	"github.com/stretchr/testify/assert"
)

var llamaCppClient *llamacpp.LlamaCppClient = &llamacpp.LlamaCppClient{
	HttpClient: http.DefaultClient,
	LlamaCppConfiguration: &llamacpp.LlamaCppConfiguration{
		HttpAddress: &netcfg.HttpAddressConfiguration{
			Host:   "127.0.0.1",
			Port:   8081,
			Scheme: "http",
		},
	},
}

func TestJsonSchemaConstrainedCompletionsAreGenerated(t *testing.T) {
	responseChannel := make(chan llamacpp.LlamaCppCompletionToken)

	go llamaCppClient.GenerateCompletion(
		responseChannel,
		llamacpp.LlamaCppCompletionRequest{
			JsonSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]string{
						"type": "string",
					},
					"temperature": map[string]string{
						"type": "integer",
					},
				},
			},
			NPredict: 100,
			Prompt: `
			User will provide the phrase. Respond with JSON matching the
			schema. Fill the schema with the infromation provided in the user
			phrase.

			---
			JSON schema:
			{
				"type": "object",
				"properties": {
					"location": { "type": "string" },
					"temperature": { "type": "integer" }
				}
			}
			---

			---
			User phrase:
			Currently there are 16 degrees in Warsaw
			---
			`,
			Stream: true,
		},
	)

	acc := ""

	for token := range responseChannel {
		if token.Error != nil {
			t.Fatal(token.Error)
		} else {
			acc += token.Content
		}
	}

	assert.Equal(t, "{ \"location\": \"Warsaw\", \"temperature\": 16 }", acc)
}
