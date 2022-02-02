package result

import "hm2/config"

type ResultInfo struct {
	Done    bool        `json:"done"`
	Message *string     `json:"message,omitempty"`
	Items   interface{} `json:"data,omitempty"`
	User    config.User `json:"-"`
}

func SetErrorResult(m string) (result ResultInfo) {
	result.Done = false
	result.Message = &m
	result.Items = nil
	return result
}
