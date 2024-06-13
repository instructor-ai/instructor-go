package instructor

const (
	DefaultMaxRetries = 3
	DefaultValidator  = false
)

type Options struct {
	Mode       *Mode
	MaxRetries *int
	validate   *bool
	// Provider specific options:
}

var defaultOptions = Options{
	Mode:       toPtr(ModeDefault),
	MaxRetries: toPtr(DefaultMaxRetries),
	validate:   toPtr(DefaultValidator),
}

func WithMode(mode Mode) Options {
	return Options{Mode: toPtr(mode)}
}

func WithMaxRetries(maxRetries int) Options {
	return Options{MaxRetries: toPtr(maxRetries)}
}

func WithValidator(validate bool) Options {
	return Options{validate: toPtr(validate)}
}

func mergeOption(old, new Options) Options {
	if new.Mode != nil {
		old.Mode = new.Mode
	}
	if new.MaxRetries != nil {
		old.MaxRetries = new.MaxRetries
	}
	if new.validate != nil {
		old.validate = new.validate
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
