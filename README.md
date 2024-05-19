# instructor-go

Instructor is a library that makes it a breeze to work with structured outputs from large language models (LLMs).

---

[![Twitter Follow](https://img.shields.io/twitter/follow/jxnlco?style=social)](https://twitter.com/jxnlco)
[![LinkedIn Follow](https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/robby-horvath/)
[![Documentation](https://img.shields.io/badge/docs-available-brightgreen)](https://go.useinstructor.com)
[![GitHub issues](https://img.shields.io/github/issues/instructor-ai/instructor-go.svg)](https://github.com/instructor-ai/instructor-go/issues)
[![Discord](https://img.shields.io/discord/1192334452110659664?label=discord)](https://discord.gg/CV8sPM5k5Y)

Built on top of [`invopop/jsonschema`](https://github.com/invopop/jsonschema) and utilizing `jsonschema` Go struct tags (so no changing code logic), it provides a simple, transparent, and user-friendly API to manage validation, retries, and streaming responses. Get ready to supercharge your LLM workflows!

## Example

As shown in the example below, by adding extra metadata to each struct field (via `jsonschema` tag) we want the model to be made aware of:

> For more information on the `jsonschema` tags available, see the [`jsonschema` godoc](https://pkg.go.dev/github.com/invopop/jsonschema?utm_source=godoc).

```go
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

	client, err := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)
	if err != nil {
		panic(err)
	}

	var person Person
	err = client.CreateChatCompletion(
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
		&person,
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
```

## Coming Soon

1. Streaming support

## Providers

Most model API providers do not provide an official Go client, so here are the ones we chose for the following providers:

- [OpenAI](https://github.com/sashabaranov/go-openai)
- [Anthropic](https://github.com/liushuangls/go-anthropic)
