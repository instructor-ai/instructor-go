package instruct

import (
	"github.com/distantmagic/paddler/llamacpp"
)

type JsonSchemaMapper struct {
	LlamaCppClient *llamacpp.LlamaCppClient
}

func (self *JsonSchemaMapper) ToJson(
	// resultChannel chan httpserver.ServerEvent,
	jsonSchema string,
) {
}
