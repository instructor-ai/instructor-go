package main

import (
	"context"
	"fmt"
	"os"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
	"github.com/instructor-ai/instructor-go/pkg/instructor"
)

type HistoricalFact struct {
	Year        int    `json:"year"         jsonschema:"title=Year of the Fact,description=Year when the fact occurred"`
	Topic       string `json:"topic"        jsonschema:"title=Topic of the Fact,description=General category or topic of the fact"`
	Description string `json:"description"  jsonschema:"title=Description of the Fact,description=Description or details of the fact"`
}

func (hf HistoricalFact) String() string {
	return fmt.Sprintf("Year: %d\nTopic: %s\nDescription: %s", hf.Year, hf.Topic, hf.Description)
}

func main() {
	ctx := context.Background()

	client := instructor.FromCohere(
		cohereclient.NewClient(cohereclient.WithToken(os.Getenv("COHERE_API_KEY"))),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)

	hfStream, err := client.ChatStream(ctx, &cohere.ChatStreamRequest{
		Model:    toPtr("command-r-plus"),
		Preamble: toPtr("Only give 3 response"),
		Message:  "Tell me about the history of artificial intelligence",
	},
		*new(HistoricalFact),
	)
	if err != nil {
		panic(err)
	}

	for instance := range hfStream {
		hf := instance.(*HistoricalFact)
		println(hf.String())
	}
}

func toPtr[T any](val T) *T {
	return &val
}
