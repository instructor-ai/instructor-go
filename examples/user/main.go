package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Person struct {
	Name string `json:"name"          jsonschema:"title=the name,description=The name of the person,example=joe,example=lucy"`
	Age  int    `json:"age,omitempty" jsonschema:"title=the age,description=The age of the person,example=25,example=67"`
}

func main() {
	ctx := context.Background()

	client, err := instructor.FromOpenAI[Person](
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)
	if err != nil {
		panic(err)
	}

	person, err := client.CreateChatCompletion(
		ctx,
		instructor.Request{
			Model: openai.GPT4Turbo20240409,
			Messages: []instructor.Message{
				{
					Role:    instructor.RoleUser,
					Content: "Extract Robby is 22 years old.",
				},
			},
		},
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
