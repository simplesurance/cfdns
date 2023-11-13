package logs_test

import (
	"testing"

	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/logs/logtotest"
)

func TestLogger(t *testing.T) {
	logger := logs.FromDriver(logtotest.ForTest(t, false), "TestLogger")

	logger.D("Debug", logs.WithString("key", "value"))
	logger.I("Info", logs.WithString("key", "value"))
	logger.W("Warn", logs.WithString("key", "value"))
	logger.E("Error", logs.WithString("key", "value"))
}
