package instructor

import (
	openai "github.com/sashabaranov/go-openai"
)

type Message = openai.ChatCompletionMessage

type Request = openai.ChatCompletionRequest
