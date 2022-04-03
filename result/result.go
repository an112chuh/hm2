package result

import "hm2/config"

type ResultInfo struct {
	Done      bool        `json:"done"`
	Message   *string     `json:"message,omitempty"`
	Items     interface{} `json:"data,omitempty"`
	Paginator Paginator   `json:"paginator,omitempty"`
	User      config.User `json:"-"`
}

type Paginator struct {
	Total     uint64 `json:"total"`
	CountPage uint   `json:"count_page"`
	Page      uint   `json:"page"`
	Offset    uint   `json:"offset"`
	Limit     uint   `json:"limit"`
}

func SetErrorResult(m string) (result ResultInfo) {
	result.Done = false
	result.Message = &m
	result.Items = nil
	return result
}
