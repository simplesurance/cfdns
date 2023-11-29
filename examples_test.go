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
		TTL:     time.Duration(30 * time.Minute),
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
