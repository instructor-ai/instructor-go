package instructor

import (
	openai "github.com/sashabaranov/go-openai"
)

type Message = openai.ChatCompletionMessage

type Request = openai.ChatCompletionRequest

type ChatMessagePart = openai.ChatMessagePart

type ChatMessageImageURL = openai.ChatMessageImageURL

type ChatMessagePartType = openai.ChatMessagePartType

const (
	ChatMessagePartTypeText     ChatMessagePartType = "text"
	ChatMessagePartTypeImageURL ChatMessagePartType = "image_url"
)
