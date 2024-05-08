package instructor

type Mode int

const (
	UnknownMode Mode = iota
	Function
	JSON
	Grammar
	GBNF = Grammar // GBNF is an alias for Grammar
)

func (m Mode) String() string {
	switch m {
	case Function:
		return "Function"
	case JSON:
		return "JSON"
	case Grammar:
		return "Grammar"
	default:
		return "Unknown Mode"
	}
}
