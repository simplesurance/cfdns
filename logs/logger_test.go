package logs_test

import (
	"io"
	"testing"
	"time"

	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/logs/logtotest"
)

func TestLogger(t *testing.T) {
	logger := logs.New(logtotest.ForTest(t, false),
		logs.WithPrefix("TestLogger"),
		logs.WithDebugEnabledFn(func() bool { return true }))

	logger.D(func(log logs.DebugFn) {
		log("Debug",
			logs.WithString("key", "value"),
			logs.WithInt("pi", 3),
			logs.WithDuration("second", time.Second),
			logs.WithError(io.EOF),
		)
	})

	logger.I("Info", logs.WithString("key", "value"))
	logger.W("Warn", logs.WithString("key", "value"))
	logger.E("Error", logs.WithString("key", "value"))
}

func TestSubLogger(t *testing.T) {
	logger := logs.New(logtotest.ForTest(t, false),
		logs.WithPrefix("TestLogger"),
		logs.WithDebugEnabledFn(func() bool { return true }))

	logger.SubLogger(
		logs.WithString("key", "value"),
		logs.WithInt("pi", 3),
		logs.WithDuration("second", time.Second),
		logs.WithError(io.EOF),
	).I("Sub")

	logger.I("Original")
}

func TestDebugEnabled(t *testing.T) {
	debugEnabled := false
	logger := logs.New(logtotest.ForTest(t, false),
		logs.WithPrefix("TestLogger"),
		logs.WithDebugEnabledFn(func() bool { return debugEnabled }))

	logger.D(func(log logs.DebugFn) {
		log("message 1")
	})

	debugEnabled = true

	logger.D(func(log logs.DebugFn) {
		log("message 2")
	})
}
