package cfdns

type cfResponseCommon struct {
	Errors     []struct{} `json:"errors"`
	Messages   []struct{} `json:"messages"`
	ResultInfo struct {
		Count      int `json:"count"`
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
}

func (rc *cfResponseCommon) setCFCommonResponse(cf *cfResponseCommon) {
	*rc = *cf
}

var _ commonResponseSetter = &cfResponseCommon{}

type commonResponseSetter interface {
	setCFCommonResponse(*cfResponseCommon)
}
