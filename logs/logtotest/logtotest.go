package logtotest

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/fatih/color"
	"golang.org/x/exp/maps"

	"github.com/simplesurance/cfdns/logs"
)

func ForTest(t *testing.T, failOnError bool) logs.Driver {
	return testerDriver{test: t, failOnError: failOnError}
}

type testerDriver struct {
	test        *testing.T
	failOnError bool
}

func (t testerDriver) GetHelper() func() {
	return t.test.Helper
}

func (t testerDriver) Send(l *logs.Entry) {
	t.test.Helper()

	msg := &bytes.Buffer{}

	fmt.Fprintf(msg, "[%s] %s", l.Severity, l.Message)

	keys := maps.Keys(l.Tags)
	slices.Sort(keys)
	for _, key := range keys {
		fmt.Fprintf(msg, "\n- %s: %v", key, format(l.Tags[key]))
	}

	logf := t.test.Log
	if t.failOnError && l.Severity == logs.Error {
		logf = t.test.Error
	}

	text := msg.String()

	switch l.Severity {
	case logs.Error:
		text = color.New(color.FgRed).Sprint(text)
	case logs.Warn:
		text = color.New(color.FgYellow).Sprint(text)
	case logs.Info:
		text = color.New(color.FgGreen).Sprint(text)
	}

	logf(text)
}

func format(v any) string {
	switch vt := v.(type) {
	case string:
		return vt
	case error:
		return fmt.Sprintf("%v", v)
	}

	return fmt.Sprintf("%v", v)
}
