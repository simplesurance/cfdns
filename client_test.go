package cfdns_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/simplesurance/cfdns"
	"github.com/simplesurance/cfdns/log"
	"github.com/simplesurance/cfdns/log/testtarget"
)

// Integration Tests
// Require an environment variable called TEST_CF_APITOKEN

const (
	envToken       = "TEST_CF_APITOKEN"
	envTestZone    = "TEST_CF_ZONE_NAME"
	testDateFormat = "2006-01-02-15-04"
)

func TestCreateCNAME(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	client, testZoneID := getClient(ctx, t)
	cname := "CNAME"
	comment := "integration test"

	// create a DNS record
	recName := testRecordName(t)
	resp, err := client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
		ZoneID:  testZoneID,
		Name:    recName,
		Type:    cname,
		Content: "github.com",
		Comment: comment,
	})
	if err != nil {
		t.Fatalf("Error creating DNS record on CloudFlare: %v", err)
	}

	defer cleanup(ctx, t, client, testZoneID, resp.ID)

	t.Logf("DNS record created with ID=%s", resp.ID)

	// assert that it is present
	var recs []*cfdns.ListRecordsResponseItem
	recs, err = cfdns.ReadAll(ctx, client.ListRecords(&cfdns.ListRecordsRequest{
		ZoneID: testZoneID,
		Name:   resp.Name,
		Type:   cname,
	}))
	if err != nil {
		t.Fatalf("Error listing DNS records: %v", err)
	}

	if len(recs) != 1 {
		t.Fatalf("Test created one record with name %q, type %q, but found %+v",
			recName, "CNAME", recs)
	}

	assertEquals(t, recName, recs[0].Name)
	assertEquals(t, cname, recs[0].Type)
	requireNotNil(t, recs[0].Proxied)
	assertEquals(t, false, recs[0].Proxied)
	assertEquals(t, comment, recs[0].Comment)
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	originalComment := "integration test"
	changedComment := "integration test"
	cname := "cname"

	ctx := t.Context()
	client, testZoneID := getClient(ctx, t)

	// create a DNS record
	recName := testRecordName(t)
	resp, err := client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
		ZoneID:  testZoneID,
		Name:    recName,
		Type:    cname,
		Content: "1.github.com",
		Comment: originalComment,
		Proxied: true,
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("Error creating DNS record on CloudFlare: %v", err)
	}

	defer cleanup(ctx, t, client, testZoneID, resp.ID)

	_, err = client.UpdateRecord(ctx, &cfdns.UpdateRecordRequest{
		ZoneID:   testZoneID,
		RecordID: resp.ID,
		Name:     recName,
		Type:     cname,
		Content:  "2.github.com",
		Comment:  changedComment,
		Proxied:  false,
		TTL:      2 * time.Hour,
	})
	if err != nil {
		t.Fatalf("Error updating DNS record on CloudFlare: %v", err)
	}

	records, err := cfdns.ReadAll(ctx, client.ListRecords(&cfdns.ListRecordsRequest{
		ZoneID: testZoneID,
		Name:   resp.Name,
		Type:   cname,
	}))
	if err != nil {
		t.Fatalf("Error list DNS record on CloudFlare: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %v", records)
	}

	assertEquals(t, "2.github.com", records[0].Content)
	assertEquals(t, changedComment, records[0].Comment)

	requireNotNil(t, records[0].TTL)
	requireNotNil(t, records[0].Proxied)

	assertEquals(t, false, records[0].Proxied)
	assertEquals(t, 2*time.Hour, records[0].TTL)
}

// Test a few cases of error to make sure error handling works.
func TestConflict(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	client, testZoneID := getClient(ctx, t)

	cases := []*struct {
		typ           string
		content       string
		wantErrorCode int
	}{
		{
			typ:           "CNAME",
			content:       "github.com",
			wantErrorCode: 81053,
		},
		{
			typ:           "A",
			content:       "1.1.1.1",
			wantErrorCode: 81058,
		},
	}

	for _, tc := range cases {
		tc := tc //nolint:copyloopvar
		t.Run(tc.typ, func(t *testing.T) {
			t.Parallel()
			comment := "integration test"

			// create a DNS record
			recName := testRecordName(t)
			resp, err := client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
				ZoneID:  testZoneID,
				Name:    recName,
				Type:    tc.typ,
				Content: tc.content,
				Comment: comment,
			})
			if err != nil {
				t.Fatalf("Error creating DNS record on CloudFlare: %v", err)
			}

			defer cleanup(ctx, t, client, testZoneID, resp.ID)

			// do it again; it must now result in a conflict
			_, err = client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
				ZoneID:  testZoneID,
				Name:    recName,
				Type:    tc.typ,
				Content: tc.content,
				Comment: comment,
			})

			var cferr cfdns.CloudFlareError
			if err == nil || !errors.As(err, &cferr) {
				t.Fatalf("Expected cfdns.CloudFlareError, got %v", err)
			}

			if !cferr.IsAnyCFErrorCode(tc.wantErrorCode) {
				t.Errorf("Expected CloudFlare error %d, got %v",
					tc.wantErrorCode, cferr)
			}
		})
	}
}

func getClient(ctx context.Context, t *testing.T) (_ *cfdns.Client, testZoneID string) {
	apitoken := os.Getenv(envToken)
	testzone := os.Getenv(envTestZone)

	if apitoken == "" {
		t.Fatalf("%v environment variable must be set with a CloudFlare API Token", envToken)
	}

	creds, err := cfdns.APIToken(apitoken)
	if err != nil {
		t.Fatal(err)
	}

	client := cfdns.NewClient(creds,
		cfdns.WithLogger(log.New(testtarget.ForTest(t, true),
			log.WithDebugEnabledFn(func() bool { return true }))),
		cfdns.WithLogSuccessfulResponses(true))

	// return the ID of the first zone
	resp := client.ListZones(&cfdns.ListZonesRequest{})
	for {
		item, err := resp.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
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

var testRecordNameRE = regexp.MustCompile(`^test-([0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{2}-[0-9]{2})-[a-z0-9]{8}.*`)

// testRecordName creates a random name to be used when creating test
// DNS records. The name encodes the date, making cleaning-up easier.
// The cleanup() function will remove test records that encode an old date.
func testRecordName(t *testing.T) string {
	testzone := os.Getenv(envTestZone)

	rnd := make([]byte, 4)
	if _, err := rand.Read(rnd); err != nil {
		t.Fatalf("Error reading random number: %v", err)
	}

	return fmt.Sprintf("test-%s-%s.%s",
		time.Now().UTC().Format(testDateFormat),
		hex.EncodeToString(rnd),
		testzone)
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
			if errors.Is(err, io.EOF) {
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

func requireNotNil(t *testing.T, v any) {
	t.Helper()
	if v == nil {
		t.Fatalf("Unexpected nil value")
	}
}

func assertEquals[T comparable](t *testing.T, want, have T) {
	t.Helper()

	if have != want {
		t.Errorf("Value does not have the expected value:\nhave: %v\nwant: %v", have, want)
	}
}
