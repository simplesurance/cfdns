# Proteus
[![Go Report Card](https://goreportcard.com/badge/github.com/simplesurance/proteus)](https://goreportcard.com/report/github.com/simplesurance/proteus)

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
