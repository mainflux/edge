package modbus

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"

	"github.com/goburrow/modbus"
	"github.com/goburrow/serial"
)

type DataPoint string

const (
	Coil            DataPoint = "coil"
	HoldingRegister DataPoint = "h_register"
	InputRegister   DataPoint = "i_register"
	Register        DataPoint = "register"
	Discrete        DataPoint = "discrete"
	FIFO            DataPoint = "fifo"
)

var (
	errInvalidInput    = errors.New("invalid input type")
	errUnsupportedRead = errors.New("invalid iotype for Write method: register")
)

type ModbusService interface {
	// Read gets data from modbus.
	Read(address, quantity uint16, iotype DataPoint) ([]byte, error)
	// Write writes a value/s on Modbus.
	Write(address, quantity uint16, value interface{}, iotype DataPoint) ([]byte, error)
	// Close closes the modbus connection.
	Close() error
}

var _ ModbusService = (*service)(nil)

// adapterService provides methods for reading and writing data on Modbus.
type service struct {
	Client  modbus.Client
	handler modbus.ClientHandler
}

// TCPHandlerOptions defines optional handler values.
type TCPHandlerOptions struct {
	Address           string         `json:"address"`
	IdleTimeout       customDuration `json:"idle_time"`
	SlaveId           byte           `json:"slave_id,omitempty"`
	Timeout           customDuration `json:"timeout,omitempty"`
	SamplingFrequency customDuration `json:"sampling_frequency,omitempty"`
}

// NewRTUClient initializes a new modbus.Client on TCP protocol from the address
// and handler options provided.
func NewTCPClient(config TCPHandlerOptions) (ModbusService, error) {
	handler := modbus.NewTCPClientHandler(config.Address)
	if !isZeroValue(config.IdleTimeout) {
		handler.IdleTimeout = config.IdleTimeout.Duration
	}
	if !isZeroValue(config.SlaveId) {
		handler.SlaveId = config.SlaveId
	}
	if !isZeroValue(config.Timeout) {
		handler.Timeout = config.Timeout.Duration
	}

	if err := handler.Connect(); err != nil {
		return nil, err
	}

	return &service{
		Client:  modbus.NewClient(handler),
		handler: handler,
	}, nil
}

// RTUHandlerOptions defines optional handler values.
type RTUHandlerOptions struct {
	Address           string             `json:"address,omitempty"`
	BaudRate          int                `json:"baud_rate,omitempty"`
	Config            serial.Config      `json:"config,omitempty"`
	DataBits          int                `json:"data_bits,omitempty"`
	IdleTimeout       customDuration     `json:"idle_timeout,omitempty"`
	Parity            string             `json:"parity,omitempty"`
	RS485             serial.RS485Config `json:"rs485,omitempty"`
	SlaveId           byte               `json:"slave_id,omitempty"`
	StopBits          int                `json:"stop_bits,omitempty"`
	Timeout           customDuration     `json:"timeout,omitempty"`
	SamplingFrequency customDuration     `json:"sampling_frequency,omitempty"`
}

// NewRTUClient initializes a new modbus.Client on RTU/ASCII protocol from the address
// and handler options provided.
func NewRTUClient(config RTUHandlerOptions) (ModbusService, error) {
	handler := modbus.NewRTUClientHandler(config.Address)
	if !isZeroValue(config.BaudRate) {
		handler.BaudRate = config.BaudRate
	}
	if !isZeroValue(config.Config) {
		handler.Config = config.Config
	}
	if !isZeroValue(config.DataBits) {
		handler.DataBits = config.DataBits
	}
	if !isZeroValue(config.IdleTimeout) {
		handler.IdleTimeout = config.IdleTimeout.Duration
	}
	if !isZeroValue(config.Parity) {
		handler.Parity = config.Parity
	}
	if !isZeroValue(config.RS485) {
		handler.RS485 = config.RS485
	}
	if !isZeroValue(config.SlaveId) {
		handler.SlaveId = config.SlaveId
	}
	if !isZeroValue(config.StopBits) {
		handler.StopBits = config.StopBits
	}
	if !isZeroValue(config.Timeout) {
		handler.Timeout = config.Timeout.Duration
	}

	if err := handler.Connect(); err != nil {
		return nil, err
	}
	return &service{
		Client: modbus.NewClient(handler),
	}, nil
}

func isZeroValue(val interface{}) bool {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

// Write writes a value/s on Modbus.
func (s *service) Write(address, quantity uint16, value interface{}, iotype DataPoint) ([]byte, error) {
	switch iotype {
	case Coil:
		switch val := value.(type) {
		case uint16:
			return s.Client.WriteSingleCoil(address, val)
		case []byte:
			return s.Client.WriteMultipleCoils(address, quantity, val)
		default:
			return nil, errInvalidInput
		}
	case Register:
		switch val := value.(type) {
		case uint16:
			return s.Client.WriteSingleRegister(address, val)
		case []byte:
			return s.Client.WriteMultipleRegisters(address, quantity, val)
		default:
			return nil, errInvalidInput
		}
	case HoldingRegister, InputRegister, Discrete, FIFO:
		return nil, fmt.Errorf("invalid iotype for Write method: %s", iotype)
	default:
		return nil, errInvalidInput
	}
}

// Read gets data from modbus.
func (s *service) Read(address uint16, quantity uint16, iotype DataPoint) ([]byte, error) {
	switch iotype {
	case Coil:
		return s.Client.ReadCoils(address, quantity)
	case Discrete:
		return s.Client.ReadDiscreteInputs(address, quantity)
	case FIFO:
		return s.Client.ReadFIFOQueue(address)
	case HoldingRegister:
		return s.Client.ReadHoldingRegisters(address, quantity)
	case InputRegister:
		return s.Client.ReadInputRegisters(address, quantity)
	case Register:
		return nil, errUnsupportedRead
	default:
		return nil, errInvalidInput
	}
}

func (s *service) Close() error {
	switch h := s.handler.(type) {
	case *modbus.RTUClientHandler:
		return h.Close()
	case *modbus.TCPClientHandler:
		return h.Close()
	default:
		return nil
	}
}

type RWOptions struct {
	ID        int          `json:"id"`
	DataPoint DataPoint    `json:"data_type"`
	Address   uint16       `json:"address"`
	Quantity  uint16       `json:"quantity"`
	Value     ValueWrapper `json:"value,omitempty"`
}

type ValueWrapper struct {
	Data interface{}
}

func (vw *ValueWrapper) UnmarshalJSON(data []byte) error {
	var num uint16
	if err := json.Unmarshal(data, &num); err == nil {
		vw.Data = num
		return nil
	}

	var byteArray []byte
	if err := json.Unmarshal(data, &byteArray); err == nil {
		vw.Data = byteArray
		return nil
	}

	return fmt.Errorf("unable to unmarshal Value")
}

type customDuration struct {
	time.Duration
}

func (cd *customDuration) UnmarshalJSON(data []byte) error {
	var durationStr string
	if err := json.Unmarshal(data, &durationStr); err != nil {
		return err
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return err
	}

	cd.Duration = duration
	return nil
}
