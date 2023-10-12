// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package main contains modbus-adapter main function to start the modbus-adapter service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v7"
	"github.com/mainflux/edge/modbus"
	"github.com/mainflux/edge/modbus/api"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"golang.org/x/sync/errgroup"
)

const svcName = "modbus"

type config struct {
	LogLevel   string `env:"MF_MODBUS_ADAPTER_LOG_LEVEL"    envDefault:"info"`
	JaegerURL  string `env:"MF_JAEGER_URL"                  envDefault:"http://localhost:14268/api/traces"`
	RPCPort    int    `env:"MF_MODBUS_ADAPTER_RPC_PORT"     envDefault:"8855"`
	RPCHost    string `env:"MF_MODBUS_ADAPTER_RPC_HOST"     envDefault:"localhost"`
	InstanceID string `env:"MF_MODBUS_ADAPTER_INSTANCE_ID"  envDefault:""`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := mflog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Fatal(fmt.Sprintf("failed to generate instanceID: %s", err))
		}
	}

	var exitCode int
	defer mflog.ExitWithError(&exitCode)

	svc := modbus.New()

	server, err := api.NewServer(svc, fmt.Sprintf("%s:%d", cfg.RPCHost, cfg.RPCPort))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	g.Go(func() error {
		return server.Start(ctx)
	})

	logger.Info(fmt.Sprintf("modbus service listening on rpc %s:%d", cfg.RPCHost, cfg.RPCPort))

	defer func() {
		if err := server.Stop(); err != nil {
			logger.Error(err.Error())
		}
	}()

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("modbus shutdown by signal: %s", sig))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("modbus service terminated: %s", err))
	}
}
