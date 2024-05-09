package modes

type Mode string

const (
	Tool         Mode = "tool_call"
	JSON         Mode = "json_mode"
	JSONSchema   Mode = "json_schema_mode"
	MarkdownJSON Mode = "markdown_json_mode"
)
