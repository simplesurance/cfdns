package logs

import (
	"maps"
	"runtime"
	"time"
)

func New(d Driver, opt ...Option) *Logger {
	ret := &Logger{
		driver:  d,
		options: opt,
	}

	ret.loadOptions()

	return ret
}

type Logger struct {
	driver         Driver
	debugEnabledFn func() bool
	options        []Option
}

func (l *Logger) SubLogger(opts ...Option) *Logger {
	options := []Option{}
	options = append(options, l.options...)
	options = append(options, opts...)

	ret := &Logger{
		driver:  l.driver,
		options: options,
	}

	ret.loadOptions()

	return ret
}

// D sends a debug log message. The design of this function is different
// from the other log levels because usually messages from this levels
// should normally be suppressed in production. The provided function
// will not be executed if logging debug messages is not explicitly
// enabled, allowing expensive debug messages that will only be payed
// if enabled.
func (l *Logger) D(lf func(log DebugFn)) {
	if !l.debugEnabledFn() {
		return
	}

	lf(l.d)
}

type DebugFn func(msg string, opt ...Option)

func (l *Logger) d(msg string, opt ...Option) {
	l.driver.GetHelper()()
	l.log(msg, Debug, opt...)
}

func (l *Logger) I(msg string, opt ...Option) {
	l.driver.GetHelper()()
	l.log(msg, Info, opt...)
}

func (l *Logger) W(msg string, opt ...Option) {
	l.driver.GetHelper()()
	l.log(msg, Warn, opt...)
}

func (l *Logger) E(msg string, opt ...Option) {
	l.driver.GetHelper()()
	l.log(msg, Error, opt...)
}

func (l *Logger) log(msg string, sev Severity, opt ...Option) {
	l.driver.GetHelper()()

	loggerOpts := applyOptions(l.options...)
	messageOpts := applyOptions(opt...)

	tags := map[string]any{}
	maps.Copy(tags, loggerOpts.Tags)
	maps.Copy(tags, messageOpts.Tags)

	msgWithPrefix := msg
	if loggerOpts.logPrefix != "" {
		msgWithPrefix = loggerOpts.logPrefix + ": " + msg
	}

	entry := &Entry{
		Timestamp: time.Now(),
		Message:   msgWithPrefix,
		Severity:  sev,
		Tags:      tags,
	}

	// include caller
	_, file, line, ok := runtime.Caller(loggerOpts.callersToSkip)
	if ok {
		entry.Caller.File = file
		entry.Caller.Line = line
	}

	l.driver.Send(entry)
}

// loadOptions apply options that change the logger. I can only be called
// before log messages are sent to it.
func (l *Logger) loadOptions() {
	settings := applyOptions(l.options...)

	l.debugEnabledFn = settings.debugEnabledFn
}

type Caller struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

func trueFn() bool {
	return true
}

func falseFn() bool {
	return false
}
