package instructor

const (
	DefaultMaxRetries = 3
	DefaultValidator  = false
)

type Options struct {
	Mode          *Mode
	MaxRetries    *int
	WithValidator *bool
	// Provider specific options:
}

var defaultOptions = Options{
	Mode:          toPtr(ModeDefault),
	MaxRetries:    toPtr(DefaultMaxRetries),
	WithValidator: toPtr(DefaultValidator),
}

func WithMode(mode Mode) Options {
	return Options{Mode: toPtr(mode)}
}

func WithMaxRetries(maxRetries int) Options {
	return Options{MaxRetries: toPtr(maxRetries)}
}

func WithValidator(withValidator bool) Options {
	return Options{WithValidator: toPtr(withValidator)}
}

func mergeOption(old, new Options) Options {
	if new.Mode != nil {
		old.Mode = new.Mode
	}
	if new.MaxRetries != nil {
		old.MaxRetries = new.MaxRetries
	}
	if new.WithValidator != nil {
		old.WithValidator = new.WithValidator
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
