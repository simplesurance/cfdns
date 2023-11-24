package cfdns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/simplesurance/cfdns/logs"
)

// CreateRecord creates a DNS record on CloudFlare.
//
// API Reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-create-dns-record
func (c *Client) CreateRecord(
	ctx context.Context,
	req *CreateRecordRequest,
) (*CreateRecordResponse, error) {
	var ttl *int
	if req.TTL != nil {
		intttl := int(req.TTL.Seconds())
		ttl = &intttl
	}

	resp, err := runWithRetry[createRecordAPIRequest, *createRecordAPIResponse](
		ctx,
		c,
		c.cfg.logger.SubLogger(logs.WithPrefix("CreateDNSRecord")),
		&request[createRecordAPIRequest]{
			method:      "POST",
			path:        fmt.Sprintf("zones/%s/dns_records", url.PathEscape(req.ZoneID)),
			queryParams: url.Values{},
			headers: http.Header{
				"content-type": {"application/json"},
			},
			body: &createRecordAPIRequest{
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

	c.cfg.logger.D(func(log logs.DebugFn) {
		log(fmt.Sprintf("Record %s %s %s created with ID=%s",
			req.Name, req.Type, req.Content, resp.body.Result.ID))
	})
	return &CreateRecordResponse{
		ID:   resp.body.Result.ID,
		Name: resp.body.Result.Name,
	}, err
}

type CreateRecordRequest struct {
	ZoneID  string
	Name    string
	Type    string
	Content string
	Proxied *bool
	Tags    *[]string
	Comment *string
	TTL     *time.Duration
}

type CreateRecordResponse struct {
	ID   string
	Name string
}

type createRecordAPIRequest struct {
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Content string    `json:"content"`
	TTL     *int      `json:"ttl,omitempty"`
	Proxied *bool     `json:"proxied,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
	Comment *string   `json:"comment,omitempty"`
}

type createRecordAPIResponse struct {
	cfResponseCommon
	Result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"result"`
}
