package logs

type Driver interface {
	Send(l *Entry)
}

type Entry struct {
	Tags     map[string]LogTag
	Message  string
	Severity Severity
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

type LogTag func()
