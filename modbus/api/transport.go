package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/edge/internal/apiutil"
	"github.com/mainflux/edge/modbus"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	protocol    = "http"
	ctSenmlJSON = "application/senml+json"
	ctSenmlCBOR = "application/senml+cbor"
	contentType = "application/json"
)

var (
	errInvalidProtocol        = errors.New("invalid protocol expect tcp/rtu")
	errInvalidReadWriteOption = errors.New("invalid read/write option")
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc modbus.ModbusService, pub messaging.Publisher, instanceID string) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	r := bone.New()
	r.Post("/:rw/:protocol/:dataPoint", otelhttp.NewHandler(kithttp.NewServer(
		readWriteEndpoint(pub),
		decodeRequest,
		encodeResponse,
		opts...,
	), "read/write"))

	r.GetFunc("/health", mainflux.Health("modbus", instanceID))
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func decodeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	ct := r.Header.Get("Content-Type")
	if ct != contentType {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	switch bone.GetValue(r, "rw") {
	case "read":
		switch bone.GetValue(r, "protocol") {
		case "tcp":
			var req readTCPRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				return nil, err
			}
			req.Datapoint = bone.GetValue(r, "dataPoint")
			return req, nil
		case "rtu":
			var req readRTURequest
			if err := json.Unmarshal(payload, &req); err != nil {
				return nil, err
			}
			req.Datapoint = bone.GetValue(r, "dataPoint")
			return req, nil
		default:
			return nil, errInvalidProtocol
		}
	case "write":
		switch bone.GetValue(r, "protocol") {
		case "tcp":
			var req writeTCPRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				return nil, err
			}
			req.Datapoint = bone.GetValue(r, "dataPoint")
			return req, nil
		case "rtu":
			var req writeRTURequest
			if err := json.Unmarshal(payload, &req); err != nil {
				return nil, err
			}
			req.Datapoint = bone.GetValue(r, "dataPoint")
			return req, nil
		default:
			return nil, errInvalidProtocol
		}
	default:
		return nil, errInvalidReadWriteOption
	}
}

func encodeResponse(_ context.Context, w http.ResponseWriter, _ interface{}) error {
	w.WriteHeader(http.StatusAccepted)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	var wrapper error
	if errors.Contains(err, apiutil.ErrValidation) {
		wrapper, err = errors.Unwrap(err)
	}

	switch {
	case errors.Contains(err, apiutil.ErrUnsupportedContentType):
		w.WriteHeader(http.StatusUnsupportedMediaType)
	case errors.Contains(err, errors.ErrMalformedEntity):
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if wrapper != nil {
		err = errors.Wrap(wrapper, err)
	}

	if errorVal, ok := err.(errors.Error); ok {
		w.Header().Set("Content-Type", contentType)

		errMsg := errorVal.Msg()
		if errorVal.Err() != nil {
			errMsg = fmt.Sprintf("%s : %s", errMsg, errorVal.Err().Msg())
		}

		if err := json.NewEncoder(w).Encode(apiutil.ErrorRes{Err: errMsg}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
