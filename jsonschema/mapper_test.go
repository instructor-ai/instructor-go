package schemer

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/distantmagic/paddler/llamacpp"
	"github.com/distantmagic/paddler/netcfg"
	// "github.com/stretchr/testify/assert"
)

var jsonSchemaMapper *JsonSchemaMapper = &JsonSchemaMapper{
	LlamaCppClient: &llamacpp.LlamaCppClient{
		HttpClient: http.DefaultClient,
		LlamaCppConfiguration: &llamacpp.LlamaCppConfiguration{
			HttpAddress: &netcfg.HttpAddressConfiguration{
				Host:   "127.0.0.1",
				Port:   8081,
				Scheme: "http",
			},
		},
	},
}

func TestJsonSchemaConstrainedCompletionsAreGenerated(t *testing.T) {
	responseChannel := make(chan JsonSchemaMapperResult)

	go jsonSchemaMapper.MapToSchema(
		responseChannel,
		map[string]any{
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
		"Currently there are 16 degrees in Warsaw",
	)

	for result := range responseChannel {
		if result.Error != nil {
			t.Fatal(result.Error)
		}

		fmt.Print(result.Result)
	}
}
