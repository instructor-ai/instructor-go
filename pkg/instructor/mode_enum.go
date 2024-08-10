package instructor

type Mode = string

const (
	ModeToolCall     Mode = "tool_call"
	ModeJSON         Mode = "json_mode"
	ModeJSONStrict   Mode = "json_strict_mode"
	ModeJSONSchema   Mode = "json_schema_mode"
	ModeMarkdownJSON Mode = "markdown_json_mode"
	ModeDefault      Mode = ModeJSONSchema
)
