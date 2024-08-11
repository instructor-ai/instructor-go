package instructor

type Mode = string

const (
	ModeToolCall          Mode = "tool_call"
	ModeJSON              Mode = "json_mode"
	ModeStructuredOutputs Mode = "structured_outputs_mode"
	ModeJSONSchema        Mode = "json_schema_mode"
	ModeMarkdownJSON      Mode = "markdown_json_mode"
	ModeDefault           Mode = ModeJSONSchema
)
