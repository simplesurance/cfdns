package textlogger

import (
	"fmt"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/fatih/color"
	"golang.org/x/exp/maps"

	"github.com/simplesurance/cfdns/log"
)

func New(out, err io.Writer) log.Driver {
	return &logger{
		outw: out,
		errw: err,
	}
}

type logger struct {
	outw io.Writer
	errw io.Writer

	outMux sync.Mutex
	errMux sync.Mutex
}

func (l *logger) Send(entry *log.Entry) {
	w := l.outw
	mux := &l.outMux
	if entry.Severity == log.Error {
		w = l.errw
		mux = &l.errMux
	}

	msg := fmt.Sprintf("%s [%s] %s",
		entry.Timestamp.Format(time.RFC3339Nano),
		entry.Severity,
		entry.Message)

	switch entry.Severity {
	case log.Info:
		msg = color.New(color.FgGreen).Sprint(msg)
	case log.Warn:
		msg = color.New(color.FgYellow).Sprint(msg)
	case log.Error:
		msg = color.New(color.FgRed).Sprint(msg)
	}

	fmt.Fprint(w, msg)

	keys := maps.Keys(entry.Tags)
	slices.Sort(keys)

	for _, key := range keys {
		val := entry.Tags[key]

		fmt.Fprint(w, color.New(color.FgMagenta).Sprintf(" %s=%v", key, val))
	}

	mux.Lock()
	defer mux.Unlock()
	fmt.Fprintln(w)
}

func (l *logger) PreLog() func() {
	return nil
}
