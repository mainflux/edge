package api

import (
	"context"
	"net"
	"net/rpc"

	"github.com/mainflux/edge/modbus"
)

type Server interface {
	Start(ctx context.Context) error
	Stop() error
}

type server struct {
	inbound *net.TCPListener
}

func NewServer(svc *modbus.Adapter, address string) (Server, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	inbound, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	if err = rpc.Register(svc); err != nil {
		return nil, err
	}
	return &server{inbound: inbound}, nil
}

func (s server) Start(ctx context.Context) error {
	rpc.Accept(s.inbound)
	return nil
}

func (s server) Stop() error {
	return s.inbound.Close()
}
