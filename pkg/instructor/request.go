package instructor

import (
	openai "github.com/sashabaranov/go-openai"
)

type (
	Message               = openai.ChatCompletionMessage
	ChatCompletionMessage = Message

	Request               = openai.ChatCompletionRequest
	ChatCompletionRequest = Request

	MessagePart     = openai.ChatMessagePart
	ChatMessagePart = MessagePart

	MessageImageURL     = openai.ChatMessageImageURL
	ChatMessageImageURL = MessageImageURL

	MessagePartType     = openai.ChatMessagePartType
	ChatMessagePartType = MessagePartType
)

const (
	MessagePartTypeText     ChatMessagePartType = "text"
	ChatMessagePartTypeText ChatMessagePartType = MessagePartTypeText

	MessagePartTypeImageURL     ChatMessagePartType = "image_url"
	ChatMessagePartTypeImageURL ChatMessagePartType = MessagePartTypeImageURL
)
