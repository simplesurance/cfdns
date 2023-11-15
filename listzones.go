package cfdns

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/simplesurance/cfdns/logs"
)

// Listzones lists zones on CloudFlare.
//
// API Reference: https://developers.cloudflare.com/api/operations/zones-get
func (c *Client) ListZones(
	ctx context.Context,
	req *ListZonesRequest,
) (*Iterator[*ListZonesResponseItem], error) {
	page := 0

	return &Iterator[*ListZonesResponseItem]{
		fetchNext: func(
			ctx context.Context,
		) ([]*ListZonesResponseItem, error) {
			queryParams := url.Values{
				"direction": {"asc"},
				"per_page":  {"100"},
			}

			if page != 0 {
				queryParams["name"] = []string{strconv.Itoa(page)}
			}

			headers := http.Header{}

			resp, err := runWithRetry[any, *listZoneAPIResponse](
				ctx,
				c.cfg.logger.SubLogger("ListZones", logs.WithInt("page", page)),
				request[any]{
					client:      c,
					method:      "GET",
					path:        "zones",
					queryParams: queryParams,
					headers:     headers,
					body:        nil,
				})
			page++

			if err != nil {
				return nil, err
			}

			items := make([]*ListZonesResponseItem, len(resp.body.Result))
			for i, v := range resp.body.Result {
				items[i] = &ListZonesResponseItem{
					ID:   v.ID,
					Name: v.Name,
				}
			}

			return items, nil
		},
	}, nil
}

type ListZonesRequest struct{}

type ListZonesResponseItem struct {
	ID   string
	Name string
}

type listZoneAPIResponse struct {
	cfResponseCommon
	Result []listZoneAPIResponseItem `json:"result"`
}

type listZoneAPIResponseItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
