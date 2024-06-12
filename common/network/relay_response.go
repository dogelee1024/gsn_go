package network

import "encoding/json"

type ApiResponse struct {
	Code    int32           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}
