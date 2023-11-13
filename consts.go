package cfdns

import (
	"time"
)

const (
	baseURL = "https://api.cloudflare.com/client/v4"

	// defaultRequestInterval specifies the minimum interval between two
	// requests sent to CloudFlare. Cloudflare by default limits clients to
	// 1200 requests every 5 minutes. The default for the client is to
	// soft-limit requests to 1000 requests / 5 minutes.
	defaultRequestInterval = time.Second * 5 * 60 / 1000
)
