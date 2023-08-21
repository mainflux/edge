package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/edge/modbus"
	"github.com/mainflux/mainflux/pkg/messaging"
)

var errInvalidRequest = errors.New("request option not supported")

func readWriteEndpoint(pub messaging.Publisher) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		switch req := request.(type) {
		case *readRTURequest:
			pubReq := struct {
				Config  modbus.RTUHandlerOptions `json:"config"`
				Options modbus.RWOptions         `json:"options"`
			}{
				Config: req.Config,
				Options: modbus.RWOptions{
					Address:  req.Address,
					Quantity: req.Qauntity,
				},
			}
			payload, err := json.Marshal(pubReq)
			if err != nil {
				return nil, err
			}

			if err = pub.Publish(ctx, fmt.Sprintf("modbus.read.rtu.%s", req.Datapoint), &messaging.Message{Payload: payload}); err != nil {
				return nil, err
			}
			return generalResponse{Payload: []byte("successful")}, nil
		case *readTCPRequest:
			pubReq := struct {
				Config  modbus.TCPHandlerOptions `json:"config"`
				Options modbus.RWOptions         `json:"options"`
			}{
				Config: req.Config,
				Options: modbus.RWOptions{
					Address:  req.Address,
					Quantity: req.Qauntity,
				},
			}
			payload, err := json.Marshal(pubReq)
			if err != nil {
				return nil, err
			}

			if err = pub.Publish(ctx, fmt.Sprintf("modbus.read.tcp.%s", req.Datapoint), &messaging.Message{Payload: payload}); err != nil {
				return nil, err
			}
			return generalResponse{Payload: []byte("successful")}, nil
		case *writeRTURequest:
			if err := req.validate(); err != nil {
				return nil, err
			}
			svc, err := modbus.NewRTUClient(req.Config)
			if err != nil {
				return nil, err
			}
			defer svc.Close()
			res, err := svc.Write(req.Address, req.Qauntity, req.Value, modbus.DataPoint(req.Datapoint))
			if err != nil {
				return nil, err
			}
			return generalResponse{Payload: res}, nil
		case *writeTCPRequest:
			if err := req.validate(); err != nil {
				return nil, err
			}
			svc, err := modbus.NewTCPClient(req.Config)
			if err != nil {
				return nil, err
			}
			defer svc.Close()
			res, err := svc.Write(req.Address, req.Qauntity, req.Value, modbus.DataPoint(req.Datapoint))
			if err != nil {
				return nil, err
			}
			return generalResponse{Payload: res}, nil
		default:
			return nil, errInvalidRequest
		}
	}
}
