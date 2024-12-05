package models

type APIResponse struct {
	Error    *APIError      `json:"error,omitempty"`
	Response *ActionConfirm `json:"response,omitempty"`
	Data     interface{}    `json:"data,omitempty"`
}

type APIError struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type ActionConfirm struct {
	Message string `json:"message"`
}
