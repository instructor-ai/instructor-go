package instructor

type MessageType int

const (
	UnknownMessageType MessageType = iota
	SystemMessage
	UserMessage
	AssistantMessage
	ToolMessage
	FunctionMessage
)

type Message struct {
	Type    MessageType
	Content string
	Role    string
	Name    string
}
