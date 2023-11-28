package log_test

import (
	"io"
	"testing"
	"time"

	"github.com/simplesurance/cfdns/log"
	"github.com/simplesurance/cfdns/log/logtotest"
)

func TestLogger(t *testing.T) {
	logger := log.New(logtotest.ForTest(t, false),
		log.WithPrefix("TestLogger"),
		log.WithDebugEnabledFn(func() bool { return true }))

	logger.D(func(lg log.DebugFn) {
		lg("Debug",
			log.WithString("key", "value"),
			log.WithInt("pi", 3),
			log.WithDuration("second", time.Second),
			log.WithError(io.EOF),
		)
	})

	logger.I("Info", log.WithString("key", "value"))
	logger.W("Warn", log.WithString("key", "value"))
	logger.E("Error", log.WithString("key", "value"))
}

func TestSubLogger(t *testing.T) {
	logger := log.New(logtotest.ForTest(t, false),
		log.WithPrefix("TestLogger"),
		log.WithDebugEnabledFn(func() bool { return true }))

	logger.SubLogger(
		log.WithString("key", "value"),
		log.WithInt("pi", 3),
		log.WithDuration("second", time.Second),
		log.WithError(io.EOF),
	).I("Sub")

	logger.I("Original")
}

func TestDebugEnabled(t *testing.T) {
	debugEnabled := false
	logger := log.New(logtotest.ForTest(t, false),
		log.WithPrefix("TestLogger"),
		log.WithDebugEnabledFn(func() bool { return debugEnabled }))

	logger.D(func(log log.DebugFn) {
		log("message 1")
	})

	debugEnabled = true

	logger.D(func(log log.DebugFn) {
		log("message 2")
	})
}
