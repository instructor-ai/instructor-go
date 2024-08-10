package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Item struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type Receipt struct {
	Items []Item  `json:"items"`
	Total float64 `json:"total"`
}

// Validate method similar to the Pydantic model validator
func (r *Receipt) Validate() error {
	calculatedTotal := 0.0
	for _, item := range r.Items {
		calculatedTotal += item.Price * float64(item.Quantity)
	}
	if calculatedTotal != r.Total {
		return fmt.Errorf("total %f does not match the sum of item prices %f", r.Total, calculatedTotal)
	}
	return nil
}

// Function to extract receipt information from a URL
func extract(ctx context.Context, client *instructor.InstructorOpenAI, url string) (*Receipt, error) {
	var receipt Receipt

	_, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o20240806,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: `Analyze the image and return the items in the receipt and the total amount.`,
				},
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf(`{"type": "image_url", "image_url": {"url": "%s"}}`, url),
				},
			},
		},
		&receipt,
	)
	if err != nil {
		return nil, err
	}

	// Validate the receipt total
	if err := receipt.Validate(); err != nil {
		return &receipt, err
	}

	return &receipt, nil
}

func main() {
	ctx := context.Background()

	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSONStrict),
		instructor.WithMaxRetries(3),
	)

	urls := []string{
		"https://templates.mediamodifier.com/645124ff36ed2f5227cbf871/supermarket-receipt-template.jpg",
		"https://ocr.space/Content/Images/receipt-ocr-original.jpg",
	}

	for _, url := range urls {
		receipt, err := extract(ctx, client, url)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		receiptJson, _ := json.MarshalIndent(receipt, "", "  ")
		fmt.Printf("Receipt: %s\n", receiptJson)
	}
}
