package logs

import "time"

type Driver interface {
	Send(l *Entry)
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
