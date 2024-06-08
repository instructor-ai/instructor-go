package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

// User contains user information
type User struct {
	FirstName      string     `validate:"required"`
	LastName       string     `validate:"required"`
	Age            uint8      `validate:"gte=0,lte=130"`
	Email          string     `validate:"required,email"`
	Gender         string     `validate:"oneof=male female prefer_not_to"`
	FavouriteColor string     `validate:"iscolor"`                // alias for 'hexcolor|rgb|rgba|hsl|hsla'
	Addresses      []*Address `validate:"required,dive,required"` // a person can have a home and cottage...
}

// Address houses a users address information
type Address struct {
	Street string `validate:"required"`
	City   string `validate:"required"`
	Planet string `validate:"required"`
	Phone  string `validate:"required"`
}

func main() {
	ctx := context.Background()
	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
		instructor.WithValidator(true),
	)

	var user User
	_, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					Content: "Extract the information from the following text: \n" +
						"Robby Smith is 22 years old, his email is robby@example.com, he is male, his favorite color is #0000FF, and his addresses are: Street: 123 Main St, City: Anytown, Planet: Earth, Phone: (555) 555-5555.",
				},
			},
		},
		&user,
	)
	if err != nil {
		log.Fatalf("Error creating chat completion: %v", err)
	}

	fmt.Printf("User Information:\n")
	fmt.Printf("First Name: %s\n", user.FirstName)
	fmt.Printf("Last Name: %s\n", user.LastName)
	fmt.Printf("Age: %d\n", user.Age)
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Gender: %s\n", user.Gender)
	fmt.Printf("Favourite Color: %s\n", user.FavouriteColor)
	fmt.Println("Addresses:")
	for _, address := range user.Addresses {
		fmt.Printf("\tStreet: %s, City: %s, Planet: %s, Phone: %s\n", address.Street, address.City, address.Planet, address.Phone)
	}
}
