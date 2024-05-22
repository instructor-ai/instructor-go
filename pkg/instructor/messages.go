package instructor

import (
	openai "github.com/sashabaranov/go-openai"
)

type (
	Message             = openai.ChatCompletionMessage
	Request             = openai.ChatCompletionRequest
	ChatMessagePart     = openai.ChatMessagePart
	ChatMessageImageURL = openai.ChatMessageImageURL
	ChatMessagePartType = openai.ChatMessagePartType
)

const (
	ChatMessagePartTypeText     ChatMessagePartType = "text"
	ChatMessagePartTypeImageURL ChatMessagePartType = "image_url"
)
