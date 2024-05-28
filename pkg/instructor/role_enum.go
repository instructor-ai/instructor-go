package instructor

type Role = string

const (
	RoleSystem            = "system"
	ChatMessageRoleSystem = RoleSystem

	RoleUser            = "user"
	ChatMessageRoleUser = RoleUser

	RoleAssistant            = "assistant"
	ChatMessageRoleAssistant = RoleAssistant

	RoleTool            = "tool"
	ChatMessageRoleTool = RoleTool
)
