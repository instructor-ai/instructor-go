package instructor_test

import ()

type Person struct {
	Name string `json:"name"          jsonschema:"title=the name,description=The name of the person,example=joe,example=lucy"`
	Age  int    `json:"age,omitempty" jsonschema:"title=the age,description=The age of the person,example=25,example=67"`
}

// func TestFromOpenAI(t *testing.T) {
// 	t.Skip("not implemented")
// 	// TODO: implement a test for FromOpenAI when it's implemented
// }
//
// type mockClient[T any] struct{}
//
// func (m *mockClient[T]) CreateCompletion(ctx context.Context, messages []Message, opts ...ClientOptions) (T, error) {
// 	return *new(T), nil
// }
//
// func (m *mockClient[T]) CreateChatCompletion(ctx context.Context, messages []Message, opts ...ClientOptions) (T, error) {
// 	return *new(T), nil
// }
//
// func TestClient(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		opts []ClientOptions
// 		want ClientOptions
// 	}{
// 		{
// 			name: "WithMode",
// 			opts: []ClientOptions{WithMode(modes.JSON)},
// 			want: ClientOptions{Mode: modes.JSON},
// 		},
// 		{
// 			name: "WithMaxRetries",
// 			opts: []ClientOptions{WithMaxRetries(5)},
// 			want: ClientOptions{MaxRetries: 5},
// 		},
// 		{
// 			name: "MultipleOptions",
// 			opts: []ClientOptions{WithMode(Function), WithMaxRetries(3)},
// 			want: ClientOptions{Mode: Function, MaxRetries: 3},
// 		},
// 	}
//
// 	mc := &mockClient[User]{}
// 	ctx := context.Background()
// 	messages := []Message{{
// 		Content: "hello",
// 		Role:    "user",
// 	}}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var opts ClientOptions
// 			for _, o := range tt.opts {
// 				opts = o
// 			}
// 			if opts != tt.want {
// 				t.Errorf("got %v, want %v", opts, tt.want)
// 			}
//
// 			_, err := mc.CreateCompletion(ctx, messages, opts)
// 			if err != nil {
// 				t.Errorf("CreateCompletion returned error: %v", err)
// 			}
//
// 			_, err = mc.CreateChatCompletion(ctx, messages, opts)
// 			if err != nil {
// 				t.Errorf("CreateChatCompletion returned error: %v", err)
// 			}
// 		})
// 	}
// }
