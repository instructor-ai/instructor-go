package main

import (
	"context"
	"fmt"
	"strings"

	cohere "github.com/cohere-ai/cohere-go/v2"
	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
)

type Section struct {
	Title      string `json:"title" jsonschema:"description=main topic of this section of the document"`
	StartIndex int    `json:"start_index" jsonschema:"description=line number where the section begins"`
	EndIndex   int    `json:"end_index" jsonschema:"description=line number where the section ends"`
}

type StructuredDocument struct {
	Sections []Section `json:"sections" jsonschema:"description=a list of sections of the document"`
}

func docWithLines(document string) (string, map[int]string) {
	documentLines := strings.Split(document, "\n")
	documentWithLineNumbers := ""
	line2text := make(map[int]string)
	for i, line := range documentLines {
		documentWithLineNumbers += fmt.Sprintf("[%d] %s\n", i, line)
		line2text[i] = line
	}
	return documentWithLineNumbers, line2text
}

// Mocking the call to the instructor and cohere client. Replace this with actual implementation.
func getStructuredDocument(documentWithLineNumbers string) StructuredDocument {
	// Mock response
	return StructuredDocument{
		Sections: []Section{
			{
				Title:      "Introduction",
				StartIndex: 0,
				EndIndex:   10,
			},
			{
				Title:      "Background",
				StartIndex: 10,
				EndIndex:   20,
			},
			// Add more sections as needed
		},
	}
}

func getSectionsText(structuredDoc StructuredDocument, line2text map[int]string) []map[string]interface{} {
	var segments []map[string]interface{}
	for _, s := range structuredDoc.Sections {
		var contents []string
		for lineID := s.StartIndex; lineID < s.EndIndex; lineID++ {
			if line, exists := line2text[lineID]; exists {
				contents = append(contents, line)
			}
		}
		segment := map[string]interface{}{
			"title":   s.Title,
			"content": strings.Join(contents, "\n"),
			"start":   s.StartIndex,
			"end":     s.EndIndex,
		}
		segments = append(segments, segment)
	}
	return segments
}

func main() {
	ctx := context.Background()

	client := cohereclient.NewClient(cohereclient.WithToken("<YOUR_AUTH_TOKEN>"))

	document := `
Introduction to Multi-Head Attention
In the very first figure, at the top of this article, we saw that transformers use a module called multi-head attention. How does that relate to the self-attention mechanism (scaled-dot product attention) we walked through above?
In the scaled dot-product attention, the input sequence was transformed using three matrices representing the query, key, and value. These three matrices can be considered as a single attention head in the context of multi-head attention. The figure below summarizes this single attention head we covered previously:
As its name implies, multi-head attention involves multiple such heads, each consisting of query, key, and value matrices. This concept is similar to the use of multiple kernels in convolutional neural networks.
To illustrate this in code, suppose we have 3 attention heads, so we now extend the \(d' \times d\) dimensional weight matrices so \(3 \times d' \times d\):
In:
h = 3
multihead_W_query = torch.nn.Parameter(torch.rand(h, d_q, d))
multihead_W_key = torch.nn.Parameter(torch.rand(h, d_k, d))
multihead_W_value = torch.nn.Parameter(torch.rand(h, d_v, d))
Consequently, each query element is now \(3 \times d_q\) dimensional, where \(d_q=24\) (here, let’s keep the focus on the 3rd element corresponding to index position 2):
In:
multihead_query_2 = multihead_W_query.matmul(x_2)
print(multihead_query_2.shape)
Out:
torch.Size([3, 24])
`

	response, err := client.Chat(ctx, &cohere.ChatRequest{
		Message: "How is the weather today?",
	},
	)
	_, _ = response, err

	documentWithLineNumbers, line2text := docWithLines(document)
	structuredDoc := getStructuredDocument(documentWithLineNumbers)
	segments := getSectionsText(structuredDoc, line2text)

	fmt.Println(segments[1]["title"])
	fmt.Println(segments[1]["content"])
}
