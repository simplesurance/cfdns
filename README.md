# CFDNS
[![Go Report Card](https://goreportcard.com/badge/github.com/simplesurance/cfdns)](https://goreportcard.com/report/github.com/simplesurance/cfdns)

## About

Non-Official GO CloudFlare DNS API client for go. It was created because
the official API is not stable and breaks both expectations multiple times
a year. Some times the change is easily detected, because causes compilation
errors, in other cases the change is not detected, like changes in the
returned error. It is designed to support only the DNS service.

## Project Status

This project is in pre-release stage and backwards compatibility is not
guaranteed.

## How to Get

```bash
go get github.com/simplesurance/cfdns@latest
```
## How to Use

### Listing Records

```go
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
```

### Create and Delete a DNS Record

```go
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
```
