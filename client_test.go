package cfdns_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/simplesurance/cfdns"
	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/logs/logtotest"
)

// Integration Tests
// Require an environment variable called TEST_CF_APITOKEN

const envToken = "TEST_CF_APITOKEN"

func TestListZones(t *testing.T) {
	ctx := context.Background()

	apitoken := os.Getenv(envToken)
	if apitoken == "" {
		t.Fatalf("%v environment variable must be set with a CloudFlare API Token", envToken)
	}

	client := cfdns.NewClient(cfdns.APIToken(apitoken),
		cfdns.WithLogger(logs.FromDriver(logtotest.ForTest(t, true), "")))

	resp, err := client.ListZones(ctx, &cfdns.ListZonesRequest{})
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	for {
		item, err := resp.Next(ctx)
		if err != nil {
			if errors.Is(err, cfdns.Done) {
				break
			}

			t.Fatalf("Error while fetching response: %v", err)
		}

		t.Logf("Found zone %s: %s", item.ID, item.Name)
	}
}
