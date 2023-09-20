package modbus

import (
	"errors"
	"sync"
)

var errUnsupportedModbusProtocol = errors.New("unsupported modbus protocol")

type Service interface {
	// Read subscribes to the Subscriber and
	// reads modbus sensor values while publishing them to publisher.
	Read(config RWOptions, res *[]byte) error
	// Write subscribes to the Subscriber and
	// writes to modbus sensor.
	Write(config RWOptions, res *[]byte) error
	// Configure sets the configuration for a device and returns the index for the connection.
	Configure(config interface{}, id *int) error
}

type service struct {
	sync.Mutex
	servers map[int]ModbusService
}

func New() Service {
	return &service{
		servers: make(map[int]ModbusService),
	}
}

func (s *service) Read(config RWOptions, res *[]byte) error {
	dat, err := s.servers[config.ID].Read(config.Address, config.Quantity, config.DataPoint)
	res = &dat
	return err
}

func (s *service) Write(config RWOptions, res *[]byte) error {
	dat, err := s.servers[config.ID].Write(config.Address, config.Quantity, config.Value.Data, config.DataPoint)
	res = &dat
	return err
}

func (s *service) Configure(config interface{}, id *int) error {
	switch conf := config.(type) {
	case TCPHandlerOptions:
		svc, err := NewTCPClient(conf)
		if err != nil {
			return err
		}
		s.Lock()
		newID := len(s.servers) + 1
		s.servers[newID] = svc
		s.Unlock()
		id = &newID
	case RTUHandlerOptions:
		svc, err := NewRTUClient(conf)
		if err != nil {
			return err
		}
		s.Lock()
		newID := len(s.servers) + 1
		s.servers[newID] = svc
		s.Unlock()
		id = &newID
	default:
		return errUnsupportedModbusProtocol
	}
	return nil
}
