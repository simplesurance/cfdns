package cfdns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/simplesurance/cfdns/logs"
)

// DeleteRecord deletes a DNS record on CloudFlare.
//
// API Reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-delete-dns-record
func (c *Client) DeleteRecord(
	ctx context.Context,
	req *DeleteRecordRequest,
) (*DeleteRecordResponse, error) {
	_, err := runWithRetry[*struct{}, *deleteRecordAPIResponse](
		ctx,
		c.cfg.logger.SubLogger(logs.WithPrefix("DeleteDNSRecord")),
		request[*struct{}]{
			client: c,
			method: "DELETE",
			path: fmt.Sprintf("zones/%s/dns_records/%s",
				url.PathEscape(req.ZoneID),
				url.PathEscape(req.RecordID)),
			queryParams: url.Values{
				"content-type": {"application/json"},
			},
			headers: http.Header{},
			body:    nil,
		})

	if err != nil {
		return nil, err
	}

	c.cfg.logger.D(func(log logs.DebugFn) {
		log(fmt.Sprintf("Record %s deleted", req.RecordID))
	})
	return &DeleteRecordResponse{}, err
}

type DeleteRecordRequest struct {
	ZoneID   string
	RecordID string
}

type DeleteRecordResponse struct{}

type deleteRecordAPIResponse struct {
	cfResponseCommon
}
