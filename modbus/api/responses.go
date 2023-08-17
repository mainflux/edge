package api

import (
	"net/http"

	"github.com/mainflux/mainflux"
)

var _ mainflux.Response = (*generalResponse)(nil)

type generalResponse struct {
	Payload []byte `json:"payload"`
}

// Code implements mainflux.Response.
func (*generalResponse) Code() int {
	return http.StatusOK
}

// Empty implements mainflux.Response.
func (*generalResponse) Empty() bool {
	return false
}

// Headers implements mainflux.Response.
func (*generalResponse) Headers() map[string]string {
	return map[string]string{}
}
