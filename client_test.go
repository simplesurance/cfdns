package cfdns_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/simplesurance/cfdns"
	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/logs/logtotest"
)

// Integration Tests
// Require an environment variable called TEST_CF_APITOKEN

const (
	envToken    = "TEST_CF_APITOKEN"
	envTestZone = "TEST_CF_ZONE_NAME"
)

// TestListZones asserts that at least one zone can be listed.
func TestListZones(t *testing.T) {
	ctx := context.Background()
	client, _ := getClient(ctx, t)

	listedZones := 0
	resp := client.ListZones(ctx, &cfdns.ListZonesRequest{})
	for {
		item, err := resp.Next(ctx)
		if err != nil {
			if errors.Is(err, cfdns.Done) {
				break
			}

			t.Fatalf("Error while fetching response: %v", err)
		}

		t.Logf("Found zone %s: %s", item.ID, item.Name)
		listedZones++

		listRecords(ctx, t, client, item)
	}

	if listedZones == 0 {
		t.Errorf("expected at least one zone to be listed")
	}
}

func TestCreateCNAME(t *testing.T) {
	ctx := context.Background()
	client, testZoneID := getClient(ctx, t)

	// create a DNS record
	recName := testRecordName(t)
	resp, err := client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
		ZoneID:  testZoneID,
		Name:    recName,
		Type:    "CNAME",
		Content: "github.com",
	})
	if err != nil {
		t.Fatalf("Error creating DNS record on CloudFlare: %v", err)
	}

	t.Logf("DNS record created with ID=%s", resp.ID)
}

// testRecordName creates a random name to be used when creating test
// DNS records. The name encodes the date, making cleaning-up easier.
func testRecordName(t *testing.T) string {
	rnd := make([]byte, 4)
	if _, err := rand.Read(rnd); err != nil {
		t.Fatalf("Error reading random number: %v", err)
	}

	return fmt.Sprintf("test-%s-%s",
		time.Now().UTC().Format("2006-01-02-15-04"),
		hex.EncodeToString(rnd))
}

func getClient(ctx context.Context, t *testing.T) (_ *cfdns.Client, testZoneID string) {
	apitoken := os.Getenv(envToken)
	testzone := os.Getenv(envTestZone)

	if apitoken == "" {
		t.Fatalf("%v environment variable must be set with a CloudFlare API Token", envToken)
	}

	client := cfdns.NewClient(cfdns.APIToken(apitoken),
		cfdns.WithLogger(logs.FromDriver(logtotest.ForTest(t, true), "")))

	// return the ID of the first zone
	resp := client.ListZones(ctx, &cfdns.ListZonesRequest{})
	for {
		item, err := resp.Next(ctx)
		if err != nil {
			if errors.Is(err, cfdns.Done) {
				break
			}

			t.Fatalf("Error while fetching response: %v", err)
		}

		if item.Name == testzone {
			return client, item.ID
		}
	}

	t.Fatalf("Zone %s not found on CloudFlare", testzone)
	return nil, ""
}

func listRecords(
	ctx context.Context,
	t *testing.T,
	client *cfdns.Client,
	item *cfdns.ListZonesResponseItem,
) {
	iter := client.ListRecords(ctx, &cfdns.ListRecordsRequest{
		ZoneID: item.ID,
	})

	totalRecords := 0
	for {
		rec, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, cfdns.Done) {
				break
			}

			t.Fatalf("Error listing DNS records from zone %s (%s): %v",
				item.Name, item.ID, err)
		}

		t.Logf("  - %s %s %s", rec.Name, rec.Type, rec.Content)
		totalRecords++
	}

	t.Logf("TOTAL DNS RECORDS: %v", totalRecords)
}
