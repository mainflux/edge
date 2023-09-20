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
	jaegerClient "github.com/mainflux/edge/internal/clients/jaeger"
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
	ServerURL  string `env:"MF_MODBUS_ADAPTER_URL"           envDefault:"http://localhost:8855"`
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

	tp, err := jaegerClient.NewProvider(svcName, cfg.JaegerURL, cfg.InstanceID)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to init Jaeger: %s", err))
	}
	var exitCode int
	defer mflog.ExitWithError(&exitCode)

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()

	svc := modbus.New()

	api.NewServer(svc, cfg.ServerURL)

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
