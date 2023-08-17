package api

import (
	"github.com/mainflux/edge/modbus"
	"github.com/mainflux/mainflux/pkg/errors"
)

type readRTURequest struct {
	Config    modbus.RTUHandlerOptions
	Address   uint16 `json:"address"`
	Qauntity  uint16 `json:"quanitty"`
	Datapoint string `json:"-"`
}

func (req readRTURequest) validate() error {
	if req.Datapoint == "" {
		return errors.ErrMalformedEntity
	}
	return nil
}

type writeRTURequest struct {
	Config    modbus.RTUHandlerOptions
	Address   uint16              `json:"address"`
	Qauntity  uint16              `json:"quanitty"`
	Value     modbus.ValueWrapper `json:"value"`
	Datapoint string              `json:"-"`
}

func (req writeRTURequest) validate() error {
	if req.Datapoint == "" {
		return errors.ErrMalformedEntity
	}
	return nil
}

type readTCPRequest struct {
	Config    modbus.TCPHandlerOptions
	Address   uint16 `json:"address"`
	Qauntity  uint16 `json:"quanitty"`
	Datapoint string `json:"-"`
}

func (req readTCPRequest) validate() error {
	if req.Datapoint == "" {
		return errors.ErrMalformedEntity
	}
	return nil
}

type writeTCPRequest struct {
	Config    modbus.TCPHandlerOptions
	Address   uint16              `json:"address"`
	Qauntity  uint16              `json:"quanitty"`
	Value     modbus.ValueWrapper `json:"value"`
	Datapoint string              `json:"-"`
}

func (req writeTCPRequest) validate() error {
	if req.Datapoint == "" {
		return errors.ErrMalformedEntity
	}
	return nil
}
