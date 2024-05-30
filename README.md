# instructor-go

Instructor is a library that makes it a breeze to work with structured outputs from large language models (LLMs).

---

[![Twitter Follow](https://img.shields.io/twitter/follow/jxnlco?style=social)](https://twitter.com/jxnlco)
[![LinkedIn Follow](https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/robby-horvath/)
[![Documentation](https://img.shields.io/badge/docs-available-brightgreen)](https://go.useinstructor.com)
[![GitHub issues](https://img.shields.io/github/issues/instructor-ai/instructor-go.svg)](https://github.com/instructor-ai/instructor-go/issues)
[![Discord](https://img.shields.io/discord/1192334452110659664?label=discord)](https://discord.gg/UD9GPjbs8c)

Built on top of [`invopop/jsonschema`](https://github.com/invopop/jsonschema) and utilizing `jsonschema` Go struct tags (so no changing code logic), it provides a simple, transparent, and user-friendly API to manage validation, retries, and streaming responses. Get ready to supercharge your LLM workflows!

## Example

As shown in the example below, by adding extra metadata to each struct field (via `jsonschema` tag) we want the model to be made aware of:

> For more information on the `jsonschema` tags available, see the [`jsonschema` godoc](https://pkg.go.dev/github.com/invopop/jsonschema?utm_source=godoc).

<details>
<summary>Running</summary>

```bash
export OPENAI_API_KEY=<Your OpenAI API Key>
go run examples/user/main.go
```

</details>

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

	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)

	var person Person
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Extract Robby is 22 years old.",
				},
			},
		},
		&person,
	)
	_ = resp // sends back original response so no information loss from original API
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

### Other Examples

<details>
<summary>Function Calling with OpenAI</summary>

<details>
<summary>Running</summary>

```bash
export OPENAI_API_KEY=<Your OpenAI API Key>
go run examples/function_calling/main.go
```

</details>

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type SearchType string

const (
	Web   SearchType = "web"
	Image SearchType = "image"
	Video SearchType = "video"
)

type Search struct {
	Topic string     `json:"topic" jsonschema:"title=Topic,description=Topic of the search,example=golang"`
	Query string     `json:"query" jsonschema:"title=Query,description=Query to search for relevant content,example=what is golang"`
	Type  SearchType `json:"type"  jsonschema:"title=Type,description=Type of search,default=web,enum=web,enum=image,enum=video"`
}

func (s *Search) execute() {
	fmt.Printf("Searching for `%s` with query `%s` using `%s`\n", s.Topic, s.Query, s.Type)
}

type Searches = []Search

func segment(ctx context.Context, data string) *Searches {

	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeToolCall),
		instructor.WithMaxRetries(3),
	)

	var searches Searches
	_, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    instructor.RoleUser,
				Content: fmt.Sprintf("Consider the data below: '\n%s' and segment it into multiple search queries", data),
			},
		},
	},
		&searches,
	)
	if err != nil {
		panic(err)
	}

	return &searches
}

func main() {
	ctx := context.Background()

	q := "Search for a picture of a cat, a video of a dog, and the taxonomy of each"
	for _, search := range *segment(ctx, q) {
		search.execute()
	}
	/*
		Searching for `cat` with query `picture of a cat` using `image`
		Searching for `dog` with query `video of a dog` using `video`
		Searching for `cat` with query `taxonomy of a cat` using `web`
		Searching for `dog` with query `taxonomy of a dog` using `web`
	*/
}
```

</details>

<details>
<summary>Text Classification with Anthropic</summary>

<details>
<summary>Running</summary>

```bash
export ANTHROPIC_API_KEY=<Your Anthropic API Key>
go run examples/classification/main.go
```

</details>

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	anthropic "github.com/liushuangls/go-anthropic/v2"
)

type LabelType string

const (
	LabelTechIssue    LabelType = "tech_issue"
	LabelBilling      LabelType = "billing"
	LabelGeneralQuery LabelType = "general_query"
)

type Label struct {
	Type LabelType `json:"type" jsonschema:"title=Label type,description=Type of the label,enum=tech_issue,enum=billing,enum=general_query"`
}

type Prediction struct {
	Labels []Label `json:"labels" jsonschema:"title=Predicted labels,description=Labels of the prediction"`
}

func classify(data string) *Prediction {
	ctx := context.Background()

	client := instructor.FromAnthropic(
		anthropic.NewClient(os.Getenv("ANTHROPIC_API_KEY")),
		instructor.WithMode(instructor.ModeToolCall),
		instructor.WithMaxRetries(3),
	)

	var prediction Prediction
	resp, err := client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Haiku20240307,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(fmt.Sprintf("Classify the following support ticket: %s", data)),
		},
		MaxTokens: 500,
	},
		&prediction,
	)
	_ = resp // sends back original response so no information loss from original API
	if err != nil {
		panic(err)
	}

	return &prediction
}

func main() {

	ticket := "My account is locked and I can't access my billing info."
	prediction := classify(ticket)

	assert(prediction.contains(LabelTechIssue), "Expected ticket to be related to tech issue")
	assert(prediction.contains(LabelBilling), "Expected ticket to be related to billing")
	assert(!prediction.contains(LabelGeneralQuery), "Expected ticket NOT to be a general query")

	fmt.Printf("%+v\n", prediction)
	/*
		&{Labels:[{Type:tech_issue} {Type:billing}]}
	*/
}

/******/

func (p *Prediction) contains(label LabelType) bool {
	for _, l := range p.Labels {
		if l.Type == label {
			return true
		}
	}
	return false
}

func assert(condition bool, message string) {
	if !condition {
		fmt.Println("Assertion failed:", message)
	}
}
```

</details>

<details>
<summary>Images with OpenAI</summary>

![List of movies](https://raw.githubusercontent.com/instructor-ai/instructor-go/main/examples/images/openai/books.png)

<details>
<summary>Running</summary>

```bash
export OPENAI_API_KEY=<Your OpenAI API Key>
go run examples/images/openai/main.go
```

</details>

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Book struct {
	Title  string  `json:"title,omitempty"  jsonschema:"title=title,description=The title of the book,example=Harry Potter and the Philosopher's Stone"`
	Author *string `json:"author,omitempty" jsonschema:"title=author,description=The author of the book,example=J.K. Rowling"`
}

type BookCatalog struct {
	Catalog []Book `json:"catalog"`
}

func (bc *BookCatalog) PrintCatalog() {
	fmt.Printf("Number of books in the catalog: %d\n\n", len(bc.Catalog))
	for _, book := range bc.Catalog {
		fmt.Printf("Title:  %s\n", book.Title)
		fmt.Printf("Author: %s\n", *book.Author)
		fmt.Println("--------------------")
	}
}

func main() {
	ctx := context.Background()

	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)

	url := "https://raw.githubusercontent.com/instructor-ai/instructor-go/main/examples/images/openai/books.png"

	var bookCatalog BookCatalog
	_, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: instructor.RoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "Extract book catelog from the image",
					},
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
		&bookCatalog,
	)

	if err != nil {
		panic(err)
	}

	bookCatalog.PrintCatalog()
	/*
		Number of books in the catalog: 15

		Title:  Pride and Prejudice
		Author: Jane Austen
		--------------------
		Title:  The Great Gatsby
		Author: F. Scott Fitzgerald
		--------------------
		Title:  The Catcher in the Rye
		Author: J. D. Salinger
		--------------------
		Title:  Don Quixote
		Author: Miguel de Cervantes
		--------------------
		Title:  One Hundred Years of Solitude
		Author: Gabriel García Márquez
		--------------------
		Title:  To Kill a Mockingbird
		Author: Harper Lee
		--------------------
		Title:  Beloved
		Author: Toni Morrison
		--------------------
		Title:  Ulysses
		Author: James Joyce
		--------------------
		Title:  Harry Potter and the Cursed Child
		Author: J.K. Rowling
		--------------------
		Title:  The Grapes of Wrath
		Author: John Steinbeck
		--------------------
		Title:  1984
		Author: George Orwell
		--------------------
		Title:  Lolita
		Author: Vladimir Nabokov
		--------------------
		Title:  Anna Karenina
		Author: Leo Tolstoy
		--------------------
		Title:  Moby-Dick
		Author: Herman Melville
		--------------------
		Title:  Wuthering Heights
		Author: Emily Brontë
		--------------------
	*/
}
```

</details>

<details>
<summary>Images with Anthropic</summary>

![List of books](https://raw.githubusercontent.com/instructor-ai/instructor-go/main/examples/images/anthropic/movies.png)

<details>
<summary>Running</summary>

```bash
export ANTHROPIC_API_KEY=<Your Anthropic API Key>
go run examples/images/anthropic/main.go
```

</details>

```go
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	"github.com/liushuangls/go-anthropic/v2"
)

type Movie struct {
	Title string `json:"title"          jsonschema:"title=title,description=The title of the movie,required=true,example=Ex Machina"`
	Year  int    `json:"year,omitempty" jsonschema:"title=year,description=The year of the movie,required=false,example=2014"`
}

type MovieCatalog struct {
	Catalog []Movie `json:"catalog"`
}

func (bc *MovieCatalog) PrintCatalog() {
	fmt.Printf("Number of movies in the catalog: %d\n\n", len(bc.Catalog))
	for _, movie := range bc.Catalog {
		fmt.Printf("Title:  %s\n", movie.Title)
		if movie.Year != 0 {
			fmt.Printf("Year:   %d\n", movie.Year)
		}
		fmt.Println("--------------------")
	}
}

func main() {
	ctx := context.Background()

	client := instructor.FromAnthropic(
		anthropic.NewClient(os.Getenv("ANTHROPIC_API_KEY")),
		instructor.WithMode(instructor.ModeJSONSchema),
		instructor.WithMaxRetries(3),
	)

	url := "https://raw.githubusercontent.com/instructor-ai/instructor-go/main/examples/images/anthropic/movies.jpg"
	data, err := urlToBase64(url)
	if err != nil {
		panic(err)
	}

	var movieCatalog MovieCatalog
	_, err = client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model: "claude-3-haiku-20240307",
		Messages: []anthropic.Message{
			{
				Role: instructor.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewImageMessageContent(anthropic.MessageContentImageSource{
						Type:      "base64",
						MediaType: "image/jpeg",
						Data:      data,
					}),
					anthropic.NewTextMessageContent("Extract the movie catalog from the screenshot"),
				},
			},
		},
		MaxTokens: 1000,
	},
		&movieCatalog,
	)
	if err != nil {
		panic(err)
	}

	movieCatalog.PrintCatalog()
	/*
		Number of movies in the catalog: 18

		Title:  Oppenheimer
		Year:   2023
		--------------------
		Title:  The Dark Knight
		Year:   2008
		--------------------
		Title:  Interstellar
		Year:   2014
		--------------------
		Title:  Inception
		Year:   2010
		--------------------
		Title:  Tenet
		Year:   2020
		--------------------
		Title:  Dunkirk
		Year:   2017
		--------------------
		Title:  Memento
		Year:   2000
		--------------------
		Title:  The Dark Knight Rises
		Year:   2012
		--------------------
		Title:  Batman Begins
		Year:   2005
		--------------------
		Title:  The Prestige
		Year:   2006
		--------------------
		Title:  Insomnia
		Year:   2002
		--------------------
		Title:  Following
		Year:   1998
		--------------------
		Title:  Man of Steel
		Year:   2013
		--------------------
		Title:  Transcendence
		Year:   2014
		--------------------
		Title:  Justice League
		Year:   2017
		--------------------
		Title:  Batman v Superman: Dawn of Justice
		Year:   2016
		--------------------
		Title:  Ending the Knight
		Year:   2016
		--------------------
		Title:  Larceny
		--------------------
	*/
}

/*
 * Image utilties
 */

func urlToBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
```

</details>

<details>
<summary>Streaming with OpenAI</summary>

<details>
<summary>Running</summary>

```bash
export OPENAI_API_KEY=<Your OpenAI API Key>
go run examples/streaming/main.go
```

</details>

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Product struct {
	ID   string `json:"product_id"   jsonschema:"title=Product ID,description=ID of the product,required=True"`
	Name string `json:"product_name" jsonschema:"title=Product Name,description=Name of the product,required=True"`
}

func (p *Product) String() string {
	return fmt.Sprintf("Product [ID: %s, Name: %s]", p.ID, p.Name)
}

type Recommendation struct {
	Product
	Reason string `json:"reason" jsonschema:"title=Recommendation Reason,description=Reason for the product recommendation"`
}

func (r *Recommendation) String() string {
	return fmt.Sprintf(`
Recommendation [
    %s
    Reason [%s]
]`, r.Product.String(), r.Reason)
}

func main() {
	ctx := context.Background()

	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
	)

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

	recommendationChan, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o20240513,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: instructor.RoleSystem,
				Content: fmt.Sprintf(`
Generate the product recommendations from the product list based on the customer profile.
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
		*new(Recommendation),
	)
	if err != nil {
		panic(err)
	}

	for instance := range recommendationChan {
		recommendation, _ := instance.(*Recommendation)
		println(recommendation.String())
	}
	/*
		Recommendation [
		    Product [ID: 7, Name: Apple MacBook Air (2023) - Latest model, high performance, portable]
		    Reason [As you have recently searched for budget laptops of 2023 and previously purchased a laptop, we believe the latest Apple MacBook Air will meet your high-performance requirements. Additionally, Apple is one of your preferred brands.]
		]

		Recommendation [
		    Product [ID: 2, Name: Apple Watch Series 7 - Advanced fitness tracking, seamless integration with Apple ecosystem]
		    Reason [Based on your recent purchase history which includes a smart watch and your preference for Apple products, we recommend the Apple Watch Series 7 for its advanced fitness tracking features.]
		]

		Recommendation [
		    Product [ID: 10, Name: Project Hail Mary by Andy Weir - Latest sci-fi book from a renowned author]
		    Reason [Given your recent search for the latest sci-fi books and frequent browsing in the Books category, 'Project Hail Mary' by Andy Weir may interest you.]
		]

		Recommendation [
		    Product [ID: 5, Name: Manduka PRO Yoga Mat - High-quality, durable, eco-friendly]
		    Reason [Since you recently searched for yoga mats and frequently browse fitness equipment, we recommend the Manduka PRO Yoga Mat to support your fitness activities.]
		]

		Recommendation [
		    Product [ID: 4, Name: AllBirds Wool Runners - Comfortable, eco-friendly sneakers]
		    Reason [Considering your preference for the AllBirds brand and your frequent browsing in fitness categories, the AllBirds Wool Runners would be a great fit for your lifestyle.]
		]
	*/
}
```

</details>

## Providers

Most model API providers do not provide an official Go client, so here are the ones we chose for the following providers:

- [OpenAI](https://github.com/sashabaranov/go-openai)
- [Anthropic](https://github.com/liushuangls/go-anthropic)
