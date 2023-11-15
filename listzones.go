package cfdns

import "context"

func (c *Client) ListZones(
	ctx context.Context,
	req *ListZonesRequest,
) (*Iterator[*ListZonesResponseItem], error) {
	return &Iterator[*ListZonesResponseItem]{
		fetchNext: func(
			ctx context.Context,
			continueToken any,
		) ([]*ListZonesResponseItem, any, error) {
			return nil, nil, nil
		},
	}, nil
}

type ListZonesRequest struct {
}

type ListZonesResponseItem struct {
}

func Next(ctx context.Context) (*DNSZone, error) {
}
