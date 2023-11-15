package logs_test

import (
	"io"
	"testing"
	"time"

	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/logs/logtotest"
)

func TestLogger(t *testing.T) {
	logger := logs.FromDriver(logtotest.ForTest(t, false), "TestLogger")

	logger.D("Debug",
		logs.WithString("key", "value"),
		logs.WithInt("pi", 3),
		logs.WithDuration("second", time.Second),
		logs.WithError(io.EOF),
	)
	logger.I("Info", logs.WithString("key", "value"))
	logger.W("Warn", logs.WithString("key", "value"))
	logger.E("Error", logs.WithString("key", "value"))
}

func TestSubLogger(t *testing.T) {
	logger := logs.FromDriver(logtotest.ForTest(t, false), "TestLogger")

	logger.SubLogger("Debug",
		logs.WithString("key", "value"),
		logs.WithInt("pi", 3),
		logs.WithDuration("second", time.Second),
		logs.WithError(io.EOF),
	).I("Sub")

	logger.I("Original")
}
