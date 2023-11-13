package logs

type options struct {
	callersToSkip int
	Tags          map[string]any
	Caller        *caller
	Severity      Severity
}

type caller struct {
	File string
	Line int
}

func applyOptions(opts ...Option) options {
	ret := options{
		callersToSkip: 1,
	}
	for _, opt := range opts {
		opt(&ret)
	}

	return ret
}

type Option func(*options)
