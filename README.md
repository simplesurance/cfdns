# CFDNS
[![Go Report Card](https://goreportcard.com/badge/github.com/simplesurance/cfdns)](https://goreportcard.com/report/github.com/simplesurance/cfdns)

## About

Non-Official GO CloudFlare DNS API client. It was created because
the official API is not stable and breaks its consumers multiple times
a year. Some of the breaks are immediately apparent because the compiler
itself can find the problem, sometimes the expectation can't be detected
automatically, while when the returned error is changed, leading to
unexpected behavior in code that might be mission-critical.

This library is being used on a system that manages dozens of domains with
thousands of records, and makes hundreds of changes per day. It show up to
be reliable, and eliminating multiple issues that were present when the
official library was being used.

The original library also has an implementation that requires an unbounded
amount of memory when listing records. This implementation is `O(1)` in
memory for all operations.

Exponential back-off is implemented and heavily tested.

This library was designed to support only the DNS service.

## Project Status

This version is stable, being used in production. It will provide backwards
compatibility according to go module conventions.

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

Rules for errors returned are as follows:

1. If the error response came from sending the HTTP response to the
   CloudFlare, even if it is an invalid response, the error is
   HTTPError;
2. If the response is a valid CloudFlare API response, then the error
   is also a CloudFlareError (so it is both HTTPError and CloudFlareError);
3. If not valid HTTP response could be obtained from the server, then some
   other error is returned.

In all cases, the caller MUST use `errors.As()` to get either the
`HTTPError` or `CloudFlareError` object.

### HTTPError

All errors that result from calling the CloudFlare REST API allow reading
the HTTP response that caused it.

```go
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
```

### CloudFlareError

```go
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
```
