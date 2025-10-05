// Package testtarget is a log driver that sends log messages to a
// go test. Specifically, log messages are sent to `logging`.T.Log.
//
// This driver can be configured to automatically fail a test if an error
// message is produced. In this case, the log message is sent to
// `logging`.T.Error.
package testtarget

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/fatih/color"

	"github.com/simplesurance/cfdns/log"
)

func ForTest(t *testing.T, failOnError bool) log.Driver {
	return testDriver{test: t, failOnError: failOnError}
}

type testDriver struct {
	test        *testing.T
	failOnError bool
}

func (t testDriver) PreLog() func() {
	return t.test.Helper
}

func (t testDriver) Send(l *log.Entry) {
	t.test.Helper()

	msg := &strings.Builder{}

	fmt.Fprintf(msg, "[%s] %s", l.Severity, l.Message)

	keys := slices.Collect(maps.Keys(l.Tags))
	slices.Sort(keys)

	for _, key := range keys {
		fmt.Fprintf(msg, "\n- %s: %v", key, format(l.Tags[key]))
	}

	logf := t.test.Log
	if t.failOnError && l.Severity == log.Error {
		logf = t.test.Error
	}

	text := msg.String()

	switch l.Severity {
	case log.Error:
		text = color.New(color.FgRed).Sprint(text)
	case log.Warn:
		text = color.New(color.FgYellow).Sprint(text)
	case log.Info:
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
