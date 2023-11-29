package cfdns_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/simplesurance/cfdns"
)

func ExampleNew() {
	ctx := context.Background()
	apitoken := os.Getenv("CFTOKEN")

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
}
