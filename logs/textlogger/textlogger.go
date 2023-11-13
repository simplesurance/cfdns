package textlogger

import (
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/simplesurance/cfdns/logs"
	"golang.org/x/exp/maps"
)

func New(out, err io.Writer) logs.Driver {
	return &logger{}
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

	fmt.Fprintf(w, "%s [%s] %s",
		entry.Timestamp.Format(time.RFC3339),
		entry.Severity,
		entry.Message)

	keys := maps.Keys(entry.Tags)
	slices.Sort(keys)

	for _, key := range keys {
		val := entry.Tags[key]

		fmt.Fprintf(w, " %s=%v", key, val)
	}

	fmt.Fprintln(w)
}

func (l *logger) GetHelper() func() {
	return func() {}
}
