package cfdns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/simplesurance/cfdns/log"
)

// DeleteRecord deletes a DNS record on CloudFlare.
//
// API Reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-delete-dns-record
func (c *Client) DeleteRecord(
	ctx context.Context,
	req *DeleteRecordRequest,
) (*DeleteRecordResponse, error) {
	_, err := sendRequestRetry[*deleteRecordAPIResponse](
		ctx,
		c,
		c.logger.SubLogger(log.WithPrefix("DeleteDNSRecord")),
		&request{
			method: http.MethodDelete,
			path: fmt.Sprintf("zones/%s/dns_records/%s",
				url.PathEscape(req.ZoneID),
				url.PathEscape(req.RecordID)),
			queryParams: url.Values{},
			body:        nil,
		})
	if err != nil {
		return nil, err
	}

	c.logger.D(func(log log.DebugFn) {
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
