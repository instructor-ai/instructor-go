package instructor

type Mode = string

const (
	ModeTool         Mode = "tool_call"
	ModeJSON         Mode = "json_mode"
	ModeJSONSchema   Mode = "json_schema_mode"
	ModeMarkdownJSON Mode = "markdown_json_mode"
	ModeDefault      Mode = ModeJSONSchema
)
