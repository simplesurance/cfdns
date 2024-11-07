package cfdns_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/simplesurance/cfdns"
)

func ExampleClient_ListZones() {
	ctx := context.Background()
	apitoken := os.Getenv("TEST_CF_APITOKEN")

	creds, err := cfdns.APIToken(apitoken)
	if err != nil {
		panic(err)
	}

	client := cfdns.NewClient(creds)

	iter := client.ListZones(&cfdns.ListZonesRequest{})
	for {
		zone, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			panic(err)
		}

		fmt.Printf("Found zone %s\n", zone.Name)
	}

	// Output: Found zone simplesurance.top
}

func ExampleClient_CreateRecord() {
	ctx := context.Background()
	apitoken := os.Getenv("TEST_CF_APITOKEN")
	testZoneID := os.Getenv("TEST_CF_ZONE_ID")

	creds, err := cfdns.APIToken(apitoken)
	if err != nil {
		panic(err)
	}

	client := cfdns.NewClient(creds)

	resp, err := client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
		ZoneID:  testZoneID,
		Name:    "example-record",
		Type:    "CNAME",
		Content: "github.com",
		Proxied: false,
		Comment: "Created by cfdns example",
		TTL:     30 * time.Minute,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created DNS record %s", resp.Name)

	// cleanup
	_, _ = client.DeleteRecord(ctx, &cfdns.DeleteRecordRequest{
		ZoneID:   testZoneID,
		RecordID: resp.ID,
	})

	// Output: Created DNS record example-record.simplesurance.top
}

func ExampleHTTPError() {
	ctx := context.Background()
	apitoken := os.Getenv("TEST_CF_APITOKEN")
	testZoneID := os.Getenv("TEST_CF_ZONE_ID")

	creds, err := cfdns.APIToken(apitoken)
	if err != nil {
		panic(err)
	}

	client := cfdns.NewClient(creds)

	_, err = client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
		ZoneID:  testZoneID,
		Name:    "invalid name",
		Type:    "A",
		Content: "github.com",
		Comment: "Created by cfdns example",
		TTL:     30 * time.Minute,
	})

	httpErr := cfdns.HTTPError{}
	if !errors.As(err, &httpErr) {
		panic("not an HTTP error")
	}

	fmt.Printf("Got HTTP error %v", httpErr.Code) // can also access response headers and raw response body

	// Output: Got HTTP error 400
}

func ExampleCloudFlareError() {
	ctx := context.Background()
	apitoken := os.Getenv("TEST_CF_APITOKEN")
	testZoneID := os.Getenv("TEST_CF_ZONE_ID")

	creds, err := cfdns.APIToken(apitoken)
	if err != nil {
		panic(err)
	}

	client := cfdns.NewClient(creds)

	_, err = client.CreateRecord(ctx, &cfdns.CreateRecordRequest{
		ZoneID:  testZoneID,
		Name:    "invalid name",
		Type:    "A",
		Content: "github.com",
		Comment: "Created by cfdns example",
		TTL:     30 * time.Minute,
	})

	cfErr := cfdns.CloudFlareError{}
	if !errors.As(err, &cfErr) {
		panic("not a CloudFlareError")
	}

	fmt.Printf("Got HTTP error %v\n", cfErr.HTTPError.Code) // can also access response headers and raw response body

	for _, cfe := range cfErr.Errors {
		fmt.Printf("- CF error %d: %s\n", cfe.Code, cfe.Message) // can also access response headers and raw response body
	}

	// Output:
	// Got HTTP error 400
	// - CF error 9005: Content for A record must be a valid IPv4 address.
}
