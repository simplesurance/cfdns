package cfdns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// UpdateRecord updates a DNS record on CloudFlare.
//
// API Reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-update-dns-record
func (c *Client) UpdateRecord(
	ctx context.Context,
	req *UpdateRecordRequest,
) (*UpdateRecordResponse, error) {
	var ttl *int
	if req.TTL != nil {
		intttl := int(req.TTL.Seconds())
		ttl = &intttl
	}

	// PUT https://api.cloudflare.com/client/v4/zones/{zone_identifier}/dns_records/{identifier}
	resp, err := runWithRetry[*updateRecordAPIRequest, *updateRecordAPIResponse](
		ctx,
		c.cfg.logger.SubLogger("UpdateDNSRecord"),
		request[*updateRecordAPIRequest]{
			client: c,
			method: "PUT",
			path: fmt.Sprintf("v4/zones/%s/dns_records/%s",
				url.PathEscape(req.ZoneID),
				url.PathEscape(req.RecordID)),
			queryParams: url.Values{},
			headers: http.Header{
				"content-type": {"application/json"},
			},
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

	c.cfg.logger.D(fmt.Sprintf("Record %s (%s %s %s) updated",
		req.Name, req.Type, req.Content, resp.body.Result.ID))
	return &UpdateRecordResponse{
		ID: resp.body.Result.ID,
	}, err
}

type UpdateRecordRequest struct {
	ZoneID   string
	RecordID string
	Name     string
	Type     string
	Content  string
	Proxied  *bool
	Tags     *[]string
	Comment  *string
	TTL      *time.Duration
}

type UpdateRecordResponse struct {
	ID string `json:"id"`
}

type updateRecordAPIRequest struct {
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Content string    `json:"content"`
	TTL     *int      `json:"ttl,omitempty"`
	Proxied *bool     `json:"proxied,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
	Comment *string   `json:"comment,omitempty"`
}

type updateRecordAPIResponse struct {
	cfResponseCommon
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}