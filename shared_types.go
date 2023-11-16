package cfdns

type cfResponseCommon struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code       int    `json:"code"`
		Message    string `json:"message"`
		ErrorChain []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error_chain"`
	} `json:"errors"`
	Messages []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"messages"`
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
