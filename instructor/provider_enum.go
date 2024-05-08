package instructor

type Provider int

const (
	UnknownProvider Provider = iota
	OpenAI
	LlamaCPP
)

func (p Provider) String() string {
	switch p {
	case OpenAI:
		return "OpenAI"
	case LlamaCPP:
		return "LlamaCPP"
	default:
		return "Unknown Provider"
	}
}
