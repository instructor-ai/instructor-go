package instructor

// import (
// 	"context"
// 	"fmt"
//
// 	cohere "github.com/cohere-ai/cohere-go/v2"
// 	cohereclient "github.com/cohere-ai/cohere-go/v2/client"
// )
//
// type CohereClient struct {
// 	Name string
//
// 	*cohereclient.Client
// }
//
// var _ Client = &CohereClient{}
//
// func NewCohereClient(client *cohereclient.Client) (*CohereClient, error) {
// 	o := &CohereClient{
// 		Name:   "Cohere",
// 		Client: client,
// 	}
// 	return o, nil
// }
//
// func (c *CohereClient) Chat(ctx context.Context, request interface{}, mode Mode, schema *Schema) (string, error) {
//
// 	req, ok := request.(cohere.ChatRequest)
// 	if !ok {
// 		return "", fmt.Errorf("invalid request type for %s client", c.Name)
// 	}
//
// 	switch mode {
// 	// case ModeToolCall:
// 	// 	return c.completionToolCall(ctx, &req, schema)
// 	case ModeJSON:
// 		return c.completionJSON(ctx, &req, schema)
// 	// case ModeJSONSchema:
// 	// 	return c.completionJSONSchema(ctx, &req, schema)
// 	default:
// 		return "", fmt.Errorf("mode '%s' is not supported for %s", mode, c.Name)
// 	}
// }
//
// func (c *CohereClient) completionJSON(ctx context.Context, request *cohere.ChatRequest, schema *Schema) (string, error) {
// 	panic("not implemented")
// }
