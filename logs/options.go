package logs

import "time"

type options struct {
	callersToSkip  int
	Tags           map[string]any
	Severity       Severity
	debugEnabledFn func() bool
	logPrefix      string
}

func applyOptions(opts ...Option) options {
	ret := options{
		callersToSkip:  2,
		Tags:           map[string]any{},
		debugEnabledFn: func() bool { return false },
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

func WithInt(key string, val int) Option {
	return func(o *options) {
		o.Tags[key] = val
	}
}

func WithDuration(key string, val time.Duration) Option {
	return func(o *options) {
		o.Tags[key] = val
	}
}

func WithError(err error) Option {
	return func(o *options) {
		o.Tags["error"] = err
	}
}

// WithPrefix can be used on a logger to configure it to include a
// prefix on all messages.
//
// If the prefix is "Foo" and a message "Bar" is send to a log, then the
// resulting log message will be "Foo: Bar".
func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.logPrefix = prefix
	}
}

// WithDebugEnabledFn allows providing a function that will be invoked
// to determine if debug log messages should be logged. Might be called
// concurrently, and must be fast.
func WithDebugEnabledFn(enabledFn func() bool) Option {
	return func(o *options) {
		o.debugEnabledFn = enabledFn
	}
}
