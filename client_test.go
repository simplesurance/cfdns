package cfdns_test

import (
	"os"

	"github.com/simplesurance/cfdns"
)

// Integration Tests
// Require an environment variable called TEST_CFTOKEN

const envToken = "TEST_CFTOKEN"

func TestListZones() {
	token := os.Getenv(envToken)

	client := cfdns.NewClient(cfdns.APIToken(envToken))
}
