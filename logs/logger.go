package logs

import "runtime"

func FromDriver(d Driver, prefix string) Logger {
	return Logger{driver: d}
}

type Logger struct {
	driver  Driver
	preffix string
}

func (l Logger) SubLogger(prefix string) Logger {
	return Logger{
		driver:  l.driver,
		preffix: prefix,
	}
}

func (l Logger) D(msg string, opt ...Option) { l.log(msg, Debug, opt...) }
func (l Logger) I(msg string, opt ...Option) { l.log(msg, Info, opt...) }
func (l Logger) W(msg string, opt ...Option) { l.log(msg, Warn, opt...) }
func (l Logger) E(msg string, opt ...Option) { l.log(msg, Error, opt...) }

func (l Logger) log(msg string, sev Severity, opt ...Option) {
	opts := applyOptions(opt...)

	entry := &Entry{
		Message:  msg,
		Severity: sev,
	}

	// include caller
	_, file, line, ok := runtime.Caller(opts.callersToSkip)
	if ok {
		opts.Tags["caller"] = Caller{
			File: file,
			Line: line,
		}
	}

	l.driver.Send(entry)
}

type Caller struct {
	File string `json:"file"`
	Line int    `json:"line"`
}
