package textlogger

import (
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/fatih/color"
	"golang.org/x/exp/maps"

	"github.com/simplesurance/cfdns/logs"
)

func New(out, err io.Writer) logs.Driver {
	return &logger{
		outw: out,
		errw: err,
	}
}

type logger struct {
	outw io.Writer
	errw io.Writer
}

func (l *logger) Send(entry *logs.Entry) {
	w := l.outw
	if entry.Severity == logs.Error {
		w = l.errw
	}

	msg := fmt.Sprintf("%s [%s] %s",
		entry.Timestamp.Format(time.RFC3339Nano),
		entry.Severity,
		entry.Message)

	switch entry.Severity {
	case logs.Info:
		msg = color.New(color.FgGreen).Sprint(msg)
	case logs.Warn:
		msg = color.New(color.FgYellow).Sprint(msg)
	case logs.Error:
		msg = color.New(color.FgRed).Sprint(msg)
	}

	fmt.Fprint(w, msg)

	keys := maps.Keys(entry.Tags)
	slices.Sort(keys)

	for _, key := range keys {
		val := entry.Tags[key]

		fmt.Fprint(w, color.New(color.FgMagenta).Sprintf(" %s=%v", key, val))
	}

	fmt.Fprintln(w)
}

func (l *logger) PreLog() func() {
	return nil
}
