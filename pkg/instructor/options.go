package instructor

const (
	DefaultMaxRetries = 3
)

type Options struct {
	Mode       *Mode
	MaxRetries *int

	// Provider specific options:
}

var defaultOptions = Options{
	Mode:       toPtr(ModeDefault),
	MaxRetries: toPtr(DefaultMaxRetries),
}

func WithMode(mode Mode) Options {
	return Options{Mode: toPtr(mode)}
}

func WithMaxRetries(maxRetries int) Options {
	return Options{MaxRetries: toPtr(maxRetries)}
}

func mergeOption(old, new Options) Options {
	if new.Mode != nil {
		old.Mode = new.Mode
	}
	if new.MaxRetries != nil {
		old.MaxRetries = new.MaxRetries
	}

	return old
}

func mergeOptions(opts ...Options) Options {
	options := defaultOptions

	for _, opt := range opts {
		options = mergeOption(options, opt)
	}

	return options
}
