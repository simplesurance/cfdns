package cfdns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/simplesurance/cfdns/logs"
)

// ListRecords lists DNS records on a zone.
//
// API Reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-list-dns-records
func (c *Client) ListRecords(
	req *ListRecordsRequest,
) *Iterator[*ListRecordsResponseItem] {
	page := 0

	return &Iterator[*ListRecordsResponseItem]{
		fetchNext: func(
			ctx context.Context,
		) ([]*ListRecordsResponseItem, bool, error) {
			page++

			queryParams := url.Values{
				"direction": {"asc"},
				"per_page":  {strconv.Itoa(itemsPerPage)},
				"page":      {strconv.Itoa(page)},
				"order":     {"type"},
			}

			if req.Name != nil {
				queryParams.Set("name", *req.Name)
			}

			if req.Type != nil {
				queryParams.Set("type", *req.Type)
			}

			resp, err := runWithRetry[*struct{}, *listRecordsAPIResponse](
				ctx,
				c.cfg.logger.SubLogger("ListRecords", logs.WithInt("page", page)),
				request[*struct{}]{
					client:      c,
					method:      "GET",
					path:        fmt.Sprintf("zones/%s/dns_records", url.PathEscape(req.ZoneID)),
					queryParams: queryParams,
					headers: http.Header{
						"content-type": {"application/json"},
					},
					body: nil,
				})

			if err != nil {
				return nil, false, err
			}

			items := make([]*ListRecordsResponseItem, len(resp.body.Result))
			for i, v := range resp.body.Result {
				items[i] = &ListRecordsResponseItem{
					ID:      v.ID,
					Name:    v.Name,
					Type:    v.Type,
					Content: v.Content,
					Proxied: v.Proxied,
				}

				if v.TTL != nil {
					duration := time.Second * time.Duration(*v.TTL)
					items[i].TTL = &duration
				}

				if v.Comment != nil {
					items[i].Comment = *v.Comment
				}
			}

			isLast := len(resp.body.Result) < itemsPerPage

			return items, isLast, nil
		},
	}
}

type ListRecordsRequest struct {
	ZoneID string
	Name   *string // Name is used to filter by name.
	Type   *string // Type is used to filter by type.
}

type ListRecordsResponseItem struct {
	ID      string
	Content string
	Name    string
	Type    string
	Proxied *bool
	Comment string
	TTL     *time.Duration
}

type listRecordsAPIResponse struct {
	cfResponseCommon
	Result []listRecordsAPIResponseItem `json:"result"`
}

type listRecordsAPIResponseItem struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Content    string    `json:"content"`
	Proxied    *bool     `json:"proxied"`
	Proxiable  bool      `json:"proxiable"`
	Type       string    `json:"type"`
	Comment    *string   `json:"comment"`
	CreatedOn  time.Time `json:"created_on"`
	ModifiedOn time.Time `json:"modified_on"`
	Locked     *bool     `json:"locked"`
	Meta       *struct {
		AutoAdded *bool   `json:"auto_added"`
		Source    *string `json:"source"`
	} `json:"meta"`
	Tags     *[]string `json:"tags"`
	TTL      *int      `json:"ttl"`
	ZoneID   *string   `json:"zone_id"`
	ZoneName string    `json:"zone_name"`
}
