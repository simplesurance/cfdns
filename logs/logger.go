package logs

import (
	"runtime"
)

func FromDriver(d Driver, prefix string) Logger {
	return Logger{driver: d, prefix: prefix}
}

type Logger struct {
	driver Driver
	prefix string
}

func (l Logger) SubLogger(prefix string) Logger {
	return Logger{
		driver: l.driver,
		prefix: prefix,
	}
}

func (l Logger) D(msg string, opt ...Option) {
	l.driver.GetHelper()()
	l.log(msg, Debug, opt...)
}

func (l Logger) I(msg string, opt ...Option) {
	l.driver.GetHelper()()
	l.log(msg, Info, opt...)
}

func (l Logger) W(msg string, opt ...Option) {
	l.driver.GetHelper()()
	l.log(msg, Warn, opt...)
}

func (l Logger) E(msg string, opt ...Option) {
	l.driver.GetHelper()()
	l.log(msg, Error, opt...)
}

func (l Logger) log(msg string, sev Severity, opt ...Option) {
	l.driver.GetHelper()()

	opts := applyOptions(opt...)

	msgWithPrefix := msg
	if l.prefix != "" {
		msgWithPrefix = l.prefix + ": " + msg
	}

	entry := &Entry{
		Message:  msgWithPrefix,
		Severity: sev,
		Tags:     map[string]any{},
	}

	// include caller
	_, file, line, ok := runtime.Caller(opts.callersToSkip)
	if ok {
		entry.Caller.File = file
		entry.Caller.Line = line
	}

	// include tags
	for k, v := range opts.Tags {
		entry.Tags[k] = v
	}

	l.driver.Send(entry)
}

type Caller struct {
	File string `json:"file"`
	Line int    `json:"line"`
}
