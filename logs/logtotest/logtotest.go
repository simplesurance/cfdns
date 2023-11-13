package logtotest

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/simplesurance/cfdns/logs"
	"golang.org/x/exp/maps"
)

func ForTest(t *testing.T) logs.Driver {
	return testerDriver{test: t}
}

type testerDriver struct {
	test *testing.T
}

func (t testerDriver) Send(l *logs.Entry) {
	msg := &bytes.Buffer{}

	fmt.Fprintf(msg, "[%s] %s", l.Severity, l.Message)

	keys := maps.Keys(l.Tags)
	slices.Sort(keys)

	for _, key := range keys {
		fmt.Fprintf(msg, "- %s: %v", key, format(l.Tags[key]))
	}

	t.test.Log(msg)
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
