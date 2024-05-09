package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	"github.com/instructor-ai/instructor-go/pkg/instructor/modes"
	openai "github.com/sashabaranov/go-openai"
)

type Person struct {
	Name string `json:"name"          jsonschema:"title=the name,description=The name of the person,example=joe,example=lucy"`
	Age  int    `json:"age,omitempty" jsonschema:"title=the age,description=The age of the person,example=25,example=67"`
}

func main() {
	ctx := context.Background()

	oai := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	client := instructor.FromOpenAI[Person](oai)

	person, err := client.CreateChatCompletion(
		ctx,
		instructor.ChatCompletionRequest{
			Model: openai.GPT4Turbo20240409,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Extract Robby is 22 years old.",
				},
			},
		},
		instructor.WithMode(modes.JSON),
		instructor.WithMaxRetries(5),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf(`
Name: %s
Age:  %d
		`, person.Name, person.Age)
	/*
		Name: Robby
		Age:  22
	*/
}
