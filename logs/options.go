package logs

type options struct {
	callersToSkip int
	Tags          map[string]any
	Severity      Severity
}

func applyOptions(opts ...Option) options {
	ret := options{
		callersToSkip: 2,
		Tags:          map[string]any{},
	}
	for _, opt := range opts {
		opt(&ret)
	}

	return ret
}

type Option func(*options)

func WithString(key, val string) Option {
	return func(o *options) {
		o.Tags[key] = val
	}
}
