package cfdns

import "slices"

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

// IsAnyCFErrorCode returns true if the CloudFlare error includes any
// of the provided codes.
func (rc *cfResponseCommon) IsAnyCFErrorCode(code ...int) bool {
	for _, haveErr := range rc.Errors {
		if slices.Contains(code, haveErr.Code) {
			return true
		}
	}

	return false
}

func (rc *cfResponseCommon) setCFCommonResponse(v *cfResponseCommon) {
	*rc = *v
}

var _ commonResponseSetter = &cfResponseCommon{}

type commonResponseSetter interface {
	setCFCommonResponse(*cfResponseCommon)
}
