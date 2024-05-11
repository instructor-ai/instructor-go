# instructor-go

_Structured large language model (LLM) extraction in Go, designed for simplicity, transparency, and control._

---

[![Twitter Follow](https://img.shields.io/twitter/follow/jxnlco?style=social)](https://twitter.com/jxnlco)
[![LinkedIn Follow](https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/robby-horvath/)
[![Documentation](https://img.shields.io/badge/docs-available-brightgreen)](https://go.useinstructor.com)
[![GitHub issues](https://img.shields.io/github/issues/instructor-ai/instructor-go.svg)](https://github.com/instructor-ai/instructor-go/issues)
[![Discord](https://img.shields.io/discord/1192334452110659664?label=discord)](https://discord.gg/CV8sPM5k5Y)

Check us out in [Python](https://python.useinstructor.com/), [JS/TS](https://js.useinstructor.com/), [Elixir](https://github.com/thmsmlr/instructor_ex/) and [PHP](https://github.com/cognesy/instructor-php/).

If you want to port Instructor to another language, please reach out to us on [Twitter](https://twitter.com/jxnlco) we'd love to help you get started!

## Usage

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

	client, err := instructor.FromOpenAI[Person](
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(5),
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
```

## Contributing

If you want to help out, checkout some of the issues marked as `good-first-issue` or `help-wanted`. Found [here](https://github.com/instructor-ai/instructor-go/labels/good%20first%20issue). They could be anything from code improvements, a guest blog post, or a new cook book.

## License

This project is licensed under the terms of the MIT License.
