package modbus

import (
	"errors"
	"sync"
)

var errDeviceNotConfigured = errors.New("modbus device is not configured")

type Service interface {
	// Read subscribes to the Subscriber and
	// reads modbus sensor values while publishing them to publisher.
	Read(config RWOptions, res *[]byte) error
	// Write subscribes to the Subscriber and
	// writes to modbus sensor.
	Write(config RWOptions, res *[]byte) error
	// ConfigureTCP sets the configuration for a TCP device and returns the index for the connection.
	ConfigureTCP(config TCPHandlerOptions, id *int) error
	// ConfigureRTU sets the configuration for a RTU/Serial device and returns the index for the connection.
	ConfigureRTU(config RTUHandlerOptions, id *int) error
	// Close closes the modbus connection.
	Close(id int, res *bool) error
}

type Adapter struct {
	mutex   sync.Mutex
	servers map[int]ModbusService
}

func New() *Adapter {
	return &Adapter{
		servers: make(map[int]ModbusService),
	}
}

func (s *Adapter) Read(config RWOptions, res *[]byte) error {
	server, ok := s.servers[config.ID]
	if !ok {
		return errDeviceNotConfigured
	}
	dat, err := server.Read(config.Address, config.Quantity, config.DataPoint)
	*res = dat
	return err
}

func (s *Adapter) Write(config RWOptions, res *[]byte) error {
	server, ok := s.servers[config.ID]
	if !ok {
		return errDeviceNotConfigured
	}
	dat, err := server.Write(config.Address, config.Quantity, config.Value.Data, config.DataPoint)
	*res = dat
	return err
}

func (s *Adapter) ConfigureTCP(config TCPHandlerOptions, id *int) error {
	svc, err := NewTCPClient(config)
	if err != nil {
		return err
	}
	s.mutex.Lock()
	newID := len(s.servers) + 1
	s.servers[newID] = svc
	s.mutex.Unlock()
	*id = newID
	return nil
}

func (s *Adapter) ConfigureRTU(config RTUHandlerOptions, id *int) error {
	svc, err := NewRTUClient(config)
	if err != nil {
		return err
	}
	s.mutex.Lock()
	newID := len(s.servers) + 1
	s.servers[newID] = svc
	s.mutex.Unlock()
	*id = newID
	return nil
}

func (s *Adapter) Close(id int, res *bool) error {
	server, ok := s.servers[id]
	if !ok {
		return errDeviceNotConfigured
	}
	if err := server.Close(); err != nil {
		return err
	}
	s.mutex.Lock()
	delete(s.servers, id)
	s.mutex.Unlock()
	*res = true
	return nil
}
