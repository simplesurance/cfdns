# CFDNS
[![Go Report Card](https://goreportcard.com/badge/github.com/simplesurance/cfdns)](https://goreportcard.com/report/github.com/simplesurance/cfdns)

## About

Non-Official GO CloudFlare DNS API client for go. It was created because
the official API is not stable and breaks its consumers multiple times
a year. Some of the breaks are immediately apparent because the compiler
itself can find the problem, sometimes the expectation can't be detected
automatically, while when the returned error is changed, leading to
unexpected behavior in code that might be mission-critical.

This library was designed to support only the DNS service.

## Project Status

This project is in pre-release stage and backwards compatibility is not
guaranteed.

## How to Get

```bash
go get github.com/simplesurance/cfdns@latest
```
## How to Use

### Listing Records

Listing records uses the _Iterator_ pattern to completely abstract the
complexity of pagination, while keeping constant memory usage, even when
the resulting list is arbitrarily large.

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

All methods that do not return a list receive a context and a request
struct and return a struct and an error.

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

## Error Handling

### CloudFlareError

### HTTPError

All errors that result from calling the CloudFlare REST API allow reading
the HTTP response that caused it.

```go

```
