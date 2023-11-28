package logs

import (
	"time"
)

type Driver interface {
	Send(l *Entry)

	// GetHelper returns a function that is called before each logging call
	// is made. This is only useful if the driver needs to find out the
	// caller and has a way of skipping some of the caller with a helper
	// function like `testing.T`.Helper(). Since the logger itself will get
	// information about the caller into log entries this is usually
	// unnecessary, so the function may return nil.
	//
	// An example of a driver that needs this helper is `logtotest`, which
	// sends log messages to a `testing.T` object, allowing the logged
	// data to be shown as part of the test, and automatically failing tests
	// that produce an error log. This loggers return `t.Helper`, and as
	// a result, the test log messages will point to the code that called
	// the logger, not to the log library itself.
	GetHelper() func()
}

type Entry struct {
	Timestamp time.Time
	Tags      map[string]any
	Message   string
	Caller    caller
	Severity  Severity
}

type caller struct {
	File string
	Line int
}

type Severity int

func (s Severity) String() string {
	return sevToString[s]
}

const (
	Debug Severity = iota
	Info
	Warn
	Error
)

var sevToString = map[Severity]string{
	Debug: "debug",
	Info:  "info",
	Warn:  "warn",
	Error: "error",
}
