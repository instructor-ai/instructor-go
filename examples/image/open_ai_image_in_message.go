package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Book struct {
	Title  string  `json:"title,omitempty" jsonschema:"title=title,description=The title of the book,example=Harry Potter and the Philosopher's Stone"`
	Author *string `json:"author,omitempty" jsonschema:"title=author,description=The author of the book,example=J.K. Rowling"`
}

type BookCatalog struct {
	Catalog []Book `json:"catalog"`
}

func (bc *BookCatalog) PrintCatalog() {
	fmt.Println("Number of books in the catalog:", len(bc.Catalog))
	for _, book := range bc.Catalog {
		fmt.Println("Title:", book.Title)
		fmt.Println("Author:", book.Author)
		fmt.Println("--------------------")
	}
}

func main() {
	ctx := context.Background()

	client, err := instructor.FromOpenAI[BookCatalog](
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(5),
	)
	if err != nil {
		panic(err)
	}

	url := "https://example.com/your-image-url.png" // Replace with your image URL

	bookCatalog, err := client.CreateChatCompletion(
		ctx,
		instructor.Request{
			Model: openai.GPT4Turbo20240409,
			Messages: []instructor.Message{
				{
					Role: instructor.RoleUser,
					MultiContent: []instructor.ChatMessagePart{
						{
							Type: instructor.ChatMessagePartTypeText,
							Text: "Extract book catelog from the image",
						},
						{
							Type: instructor.ChatMessagePartTypeImageURL,
							ImageURL: &instructor.ChatMessageImageURL{
								URL: url,
							},
						},
					},
				},
			},
		},
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Number of books in the catalog:", len(bookCatalog.Catalog))
	bookCatalog.PrintCatalog()
}
