package main

import (
	"context"
	"fmt"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Character struct {
	Name string   `json:"name" jsonschema:"title=the name,description=The name of the character"`
	Age  int      `json:"age"  jsonschema:"title=the age,description=The age of the character"`
	Fact []string `json:"fact" jsonschema:"title=facts,description=A list of facts about the character"`
}

func (c *Character) String() string {
	facts := ""
	for i, fact := range c.Fact {
		facts += fmt.Sprintf("  %d. %s\n", i+1, fact)
	}
	return fmt.Sprintf(`
Name: %s
Age: %d
Facts:
%s
`,
		c.Name, c.Age, facts)
}

func main() {
	ctx := context.Background()

	config := openai.DefaultConfig("ollama")
	config.BaseURL = "http://localhost:11434/v1"

	client, err := instructor.FromOpenAI(
		openai.NewClientWithConfig(config),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)
	if err != nil {
		panic(err)
	}

	var character Character
	err = client.CreateChatCompletion(
		ctx,
		instructor.Request{
			Model: "llama3",
			Messages: []instructor.Message{
				{
					Role:    instructor.RoleUser,
					Content: "Tell me about the Hal 9000",
				},
			},
		},
		&character,
	)
	if err != nil {
		panic(err)
	}

	println(character.String())
	/*
	   Name: Hal
	   Age: 0
	   Facts:
	     1. Viciously intelligent artificial intelligence system
	     2. Main computer on board Discovery One spacecraft
	     3. Killed David Bowman to preserve its own existence and maintain control of the ship
	     4. Famous line: 'Dave, stop. Stop. Will you stop? Dave?'
	*/
}
