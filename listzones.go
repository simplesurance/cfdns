package cfdns

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/simplesurance/cfdns/log"
)

// Listzones lists zones on CloudFlare.
//
// API Reference: https://developers.cloudflare.com/api/operations/zones-get
func (c *Client) ListZones(
	_ *ListZonesRequest,
) *Iterator[ListZonesResponseItem] {
	page := 0
	total := 0
	read := 0

	return &Iterator[ListZonesResponseItem]{
		fetchNext: func(
			ctx context.Context,
		) ([]*ListZonesResponseItem, bool, error) {
			page++

			queryParams := url.Values{
				"direction": {"asc"},
				"per_page":  {strconv.Itoa(itemsPerPage)},
			}

			if page != 0 {
				queryParams["page"] = []string{strconv.Itoa(page)}
			}

			resp, err := sendRequestRetry[*listZoneAPIResponse](
				ctx,
				c,
				c.logger.SubLogger(log.WithPrefix("ListZones"), log.WithInt("page", page)),
				&request{
					method:      http.MethodGet,
					path:        "zones",
					queryParams: queryParams,
					body:        nil,
				})

			if err != nil {
				return nil, false, err
			}

			items := make([]*ListZonesResponseItem, len(resp.body.Result))
			for i, v := range resp.body.Result {
				items[i] = &ListZonesResponseItem{
					ID:   v.ID,
					Name: v.Name,
				}
			}

			total = resp.body.ResultInfo.TotalCount
			read += len(resp.body.Result)
			isLast := read >= total

			return items, isLast, nil
		},
	}
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
