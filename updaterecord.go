package cfdns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/simplesurance/cfdns/log"
)

// UpdateRecord updates a DNS record on CloudFlare. A TTL of 1 second or less
// will use the "automatic" TTL from CloudFlare.
//
// API Reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-update-dns-record
func (c *Client) UpdateRecord(
	ctx context.Context,
	req *UpdateRecordRequest,
) (*UpdateRecordResponse, error) {
	ttl := 1 // on CF the value "1" means "automatic"
	if req.TTL > time.Second {
		ttl = int(req.TTL.Seconds())
	}

	// PUT https://api.cloudflare.com/client/v4/zones/{zone_identifier}/dns_records/{identifier}
	resp, err := sendRequestRetry[*updateRecordAPIResponse](
		ctx,
		c,
		c.logger.SubLogger(log.WithPrefix("UpdateDNSRecord")),
		&request{
			method: http.MethodPut,
			path: fmt.Sprintf("zones/%s/dns_records/%s",
				url.PathEscape(req.ZoneID),
				url.PathEscape(req.RecordID)),
			queryParams: url.Values{},
			body: &updateRecordAPIRequest{
				Name:    req.Name,
				Type:    req.Type,
				Content: req.Content,
				Proxied: req.Proxied,
				Tags:    req.Tags,
				Comment: req.Comment,
				TTL:     ttl,
			},
		})

	if err != nil {
		return nil, err
	}

	c.logger.D(func(log log.DebugFn) {
		log(fmt.Sprintf("Record %s (%s %s %s) updated",
			req.RecordID, req.Name, req.Type, req.Content))
	})
	return &UpdateRecordResponse{
		ModifiedOn: resp.body.Result.ModifiedOn,
	}, err
}

type UpdateRecordRequest struct {
	ZoneID   string
	RecordID string
	Name     string
	Type     string
	Content  string
	Proxied  bool
	Tags     []string
	Comment  string
	TTL      time.Duration
}

type UpdateRecordResponse struct {
	ModifiedOn time.Time
}

type updateRecordAPIRequest struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Content string   `json:"content"`
	TTL     int      `json:"ttl,omitempty"`
	Proxied bool     `json:"proxied,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Comment string   `json:"comment,omitempty"`
}

type updateRecordAPIResponse struct {
	cfResponseCommon
	Result struct {
		ModifiedOn time.Time `json:"modified_on"`
	} `json:"result"`
}
