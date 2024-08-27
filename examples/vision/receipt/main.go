package main

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Item struct {
	Name  string  `json:"name"     jsonschema:"title=Item Name,description=The name of the item,example=Apple,example=Banana"`
	Price float64 `json:"price"    jsonschema:"title=Item Price,description=The price of the item in dollars,example=1.99,example=2.50"`
}

func (i Item) String() string {
	return fmt.Sprintf("  Item: %s, Price: $%.2f", i.Name, i.Price)
}

type Receipt struct {
	Items []Item  `json:"items" jsonschema:"title=Receipt Items,description=The list of items in the receipt"`
	Total float64 `json:"total" jsonschema:"title=Receipt Total,description=The total cost of all items in the receipt,example=10.99,example=25.50"`
}

func (r Receipt) String() string {
	var result string
	for _, item := range r.Items {
		result += item.String() + "\n"
	}
	result += fmt.Sprintf("Total: $%.2f", r.Total)
	return result
}

func (r *Receipt) Validate() error {
	calculatedTotal := 0.0
	for _, item := range r.Items {
		calculatedTotal += item.Price
	}

	calculatedTotal = math.Round(calculatedTotal*10) / 10
	expectedTotal := math.Round(r.Total*10) / 10

	if calculatedTotal != expectedTotal {
		return fmt.Errorf("total %f does not match the sum of item prices %f", r.Total, calculatedTotal)
	}
	return nil
}

func extract(ctx context.Context, client *instructor.InstructorOpenAI, url string) (*Receipt, error) {

	var receipt Receipt
	_, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: `Analyze the image and return the items (include tax and coupons as their own items) in the receipt and the total amount.`,
				},
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: url,
							},
						},
					},
				},
			},
		},
		&receipt,
	)
	if err != nil {
		return nil, err
	}

	if err := receipt.Validate(); err != nil {
		return &receipt, err
	}

	return &receipt, nil
}

func main() {
	ctx := context.Background()

	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)

	urls := []string{
		// source: https://templates.mediamodifier.com/645124ff36ed2f5227cbf871/supermarket-receipt-template.jpg
		"https://raw.githubusercontent.com/instructor-ai/instructor-go/main/examples/vision/receipt/supermarket-receipt-template.jpg",
		// source: https://ocr.space/Content/Images/receipt-ocr-original.jpg
		"https://raw.githubusercontent.com/instructor-ai/instructor-go/main/examples/vision/receipt/receipt-ocr-original.jpg",
	}

	for _, url := range urls {
		receipt, err := extract(ctx, client, url)
		println("Receipt:\n%s\n", receipt)
		if err != nil {
			println("Error: %v\n", err)
			continue
		}
		println("\n--------------------------------\n")
	}
	/*
	   receipt:

	   	item: lorem ipsum, price: $9.20
	   	item: lorem ipsum dolor sit, price: $19.20
	   	item: lorem ipsum dolor sit amet, price: $15.00
	   	item: lorem ipsum, price: $15.00
	   	item: lorem ipsum, price: $15.00
	   	item: lorem ipsum dolor sit, price: $15.00
	   	item: lorem ipsum, price: $19.20

	   total: $107.60

	   --------------------------------

	   receipt:

	   	item: pet toy, price: $1.97
	   	item: floppy puppy, price: $1.97
	   	item: sssupreme s, price: $4.97
	   	item: 2.5 squeak, price: $5.92
	   	item: munchy dmbel, price: $3.77
	   	item: dog treat, price: $2.92
	   	item: ped pch 1, price: $0.50
	   	item: ped pch 1, price: $0.50
	   	item: hnymd smores, price: $3.98
	   	item: french drsng, price: $1.98
	   	item: 3 oranges, price: $5.47
	   	item: baby carrots, price: $1.48
	   	item: collards, price: $1.24
	   	item: calzone, price: $2.50
	   	item: mm rvw mnt, price: $19.77
	   	item: stkobrlpliabl, price: $1.97
	   	item: stkobrlpliabl, price: $1.97
	   	item: stko sunflwr, price: $0.97
	   	item: stko sunflwr, price: $0.97
	   	item: stko sunflwr, price: $0.97
	   	item: stko sunflwr, price: $0.97
	   	item: bling beads, price: $0.97
	   	item: great value, price: $9.97
	   	item: lipton, price: $4.44
	   	item: dry dog, price: $12.44
	   	item: coupon 2310652, price: $-1.00
	   	item: tax, price: $4.59

	   total: $98.21
	*/
}
