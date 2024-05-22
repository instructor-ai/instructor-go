package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Product struct {
	ID   string `json:"product_id"   jsonschema:"title=Product ID,description=ID of the product,required=True"`
	Name string `json:"product_name" jsonschema:"title=Product Name,description=Name of the product,required=True"`
}

func (p *Product) String() string {
	return fmt.Sprintf("product=[id=%s,name=%s]", p.ID, p.Name)
}

type Recommendation struct {
	Product
	Reason string `json:"reason" jsonschema:"title=Recommendation Reason,description=Reason for the product recommendation"`
}

func (r *Recommendation) String() string {
	return fmt.Sprintf("recommendation=[%s, reason=%s]", r.Product.String(), r.Reason)
}

type Recommendations struct {
	Items []Recommendation `json:"items" jsonschema:"title=Product Recommendations,description=List of product recommendations"`
}

func main() {
	ctx := context.Background()

	client, err := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
	)
	if err != nil {
		panic(err)
	}

	profileData := `
Customer ID: 12345
Recent Purchases: [Laptop, Wireless Headphones, Smart Watch]
Frequently Browsed Categories: [Electronics, Books, Fitness Equipment]
Product Ratings: {Laptop: 5 stars, Wireless Headphones: 4 stars}
Recent Search History: [best budget laptops 2023, latest sci-fi books, yoga mats]
Preferred Brands: [Apple, AllBirds, Bench]
Responses to Previous Recommendations: {Philips: Not Interested, Adidas: Not Interested}
Loyalty Program Status: Gold Member
Average Monthly Spend: $500
Preferred Shopping Times: Weekend Evenings
...
`

	products := []Product{
		{ID: "1", Name: "Sony WH-1000XM4 Wireless Headphones - Noise-canceling, long battery life"},
		{ID: "2", Name: "Apple Watch Series 7 - Advanced fitness tracking, seamless integration with Apple ecosystem"},
		{ID: "3", Name: "Kindle Oasis - Premium e-reader with adjustable warm light"},
		{ID: "4", Name: "AllBirds Wool Runners - Comfortable, eco-friendly sneakers"},
		{ID: "5", Name: "Manduka PRO Yoga Mat - High-quality, durable, eco-friendly"},
		{ID: "6", Name: "Bench Hooded Jacket - Stylish, durable, suitable for outdoor activities"},
		{ID: "7", Name: "Apple MacBook Air (2023) - Latest model, high performance, portable"},
		{ID: "8", Name: "GoPro HERO9 Black - 5K video, waterproof, for action photography"},
		{ID: "9", Name: "Nespresso Vertuo Next Coffee Machine - Quality coffee, easy to use, compact design"},
		{ID: "10", Name: "Project Hail Mary by Andy Weir - Latest sci-fi book from a renowned author"},
	}

	productList := ""
	for _, product := range products {
		productList += product.String() + "\n"
	}

	recommendationChan, err := client.CreateChatCompletionStream(
		ctx,
		instructor.Request{
			Model: openai.GPT4o20240513,
			Messages: []instructor.Message{
				{
					Role: instructor.RoleSystem,
					Content: fmt.Sprintf(`Generate the top 3 product recommendations from the product list based on the customer profile.
Return in order of highest recommended first.
Product list:
%s`, productList),
				},
				{
					Role:    instructor.RoleUser,
					Content: fmt.Sprintf("User profile:\n%s", profileData),
				},
			},
			Stream: true,
		},
		new(Recommendations),
		*new(Recommendation),
	)
	if err != nil {
		panic(err)
	}

	for instance := range recommendationChan {
		recommendation, ok := instance.(*Recommendation)
		if !ok {
			// Handle error: the received value is not a *Recommendations
			println("channel is not of correct type. Actual type: " + reflect.TypeOf(instance).String())
			continue
		}

		println(recommendation.String())
	}

	// for instance := range recommendationsChan {
	// 	recommendations, ok := instance.(*Recommendations)
	// 	if !ok {
	// 		// Handle error: the received value is not a *Recommendations
	// 		continue
	// 	}
	//
	// 	for _, recommendation := range recommendations.Items {
	// 		println(recommendation.String())
	// 	}
	// }
}
