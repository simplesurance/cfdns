package cfdns_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/simplesurance/cfdns"
	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/logs/logtotest"
)

// Integration Tests
// Require an environment variable called TEST_CF_APITOKEN

const (
	envToken       = "TEST_CF_APITOKEN"
	envTestZone    = "TEST_CF_ZONE_NAME"
	testDateFormat = "2006-01-02-15-04"
)

// TestListZones asserts that at least one zone can be listed.
func TestListZones(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client, _ := getClient(ctx, t)

	listedZones := 0
	resp := client.ListZones(&cfdns.ListZonesRequest{})
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
	t.Parallel()
	ctx := context.Background()
	client, testZoneID := getClient(ctx, t)
	cname := "CNAME"

	// create a DNS record
	recName := testRecordName(t)
	resp, err := client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
		ZoneID:  testZoneID,
		Name:    recName,
		Type:    cname,
		Content: "github.com",
	})
	if err != nil {
		t.Fatalf("Error creating DNS record on CloudFlare: %v", err)
	}

	defer cleanup(ctx, t, client, testZoneID, resp.ID)

	t.Logf("DNS record created with ID=%s", resp.ID)

	// assert that it is present
	recs, err := cfdns.ReadAll(ctx, client.ListRecords(&cfdns.ListRecordsRequest{
		ZoneID: testZoneID,
		Name:   &resp.Name,
		Type:   &cname,
	}))

	if len(recs) != 1 {
		t.Fatalf("Test created one record with name %q, type %q, but found %+v",
			recName, "CNAME", recs)
	}

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
	resp := client.ListZones(&cfdns.ListZonesRequest{})
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
	iter := client.ListRecords(&cfdns.ListRecordsRequest{
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

var testRecordNameRE = regexp.MustCompile(`^test-([0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{2}-[0-9]{2})-[a-z0-9]{8}.*`)

// testRecordName creates a random name to be used when creating test
// DNS records. The name encodes the date, making cleaning-up easier.
// The cleanup() function will remove test records that encode an old date.
func testRecordName(t *testing.T) string {
	rnd := make([]byte, 4)
	if _, err := rand.Read(rnd); err != nil {
		t.Fatalf("Error reading random number: %v", err)
	}

	return fmt.Sprintf("test-%s-%s",
		time.Now().UTC().Format(testDateFormat),
		hex.EncodeToString(rnd))
}

// cleanup removes the records with the provided ID and removes all records
// with a name that matches the names returned by testRecordName that are
// old.
func cleanup(
	ctx context.Context,
	t *testing.T,
	client *cfdns.Client,
	zoneID string,
	recIDs ...string,
) {
	// remove the records explicitly specified
	for _, recID := range recIDs {
		_, err := client.DeleteRecord(ctx, &cfdns.DeleteRecordRequest{
			ZoneID:   zoneID,
			RecordID: recID,
		})
		if err != nil {
			t.Errorf("Error removing record %s from zone %s",
				recID, zoneID)
		}
	}

	// search for old records
	iter := client.ListRecords(&cfdns.ListRecordsRequest{ZoneID: zoneID})
	for {
		record, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, cfdns.Done) {
				break
			}

			t.Logf("Error listing records when looking for old test data: %v", err)
			return
		}

		matches := testRecordNameRE.FindStringSubmatch(record.Name)
		if matches == nil {
			continue
		}

		createdOn, err := time.Parse(testDateFormat, matches[1])
		if err != nil {
			t.Errorf("Record %s (%s %s %s) has a time part %q that is invalid",
				record.ID, record.Name, record.Type, record.Content, matches[1])
			continue
		}

		if time.Since(createdOn) < time.Hour {
			break // test record created shortly ago; leave it alone
		}

		// the record is a leftover from previous test runs; remove it
		_, err = client.DeleteRecord(ctx, &cfdns.DeleteRecordRequest{
			ZoneID:   zoneID,
			RecordID: record.ID,
		})
		if err != nil {
			// only log errors
			t.Logf("WARN: error cleaning-up leftovers from previous run. Deleting record %s %s %s (%s) failed: %v",
				record.ID, record.Name, record.Type, record.Content, err)
		}
	}
}
