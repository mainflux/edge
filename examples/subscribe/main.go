package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mainflux/mainflux/pkg/errors"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	var showHelp = flag.Bool("h", false, "Show help message")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		showUsageAndExit(0)
	}

	args := flag.Args()
	if len(args) != 2 {
		showUsageAndExit(1)
	}

	subj, format := args[0], args[1]

	logger, err := mflog.New(os.Stdout, "info")
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}

	ps, err := brokers.NewPubSub(*urls, "", logger)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer ps.Close()

	handler := handler{logger: logger, format: format}

	if err := ps.Subscribe("edge", fmt.Sprintf("channels.%s", subj), &handler); err != nil {
		logger.Error(err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("subscriber shutdown by signal: %s", sig))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("subscriber terminated: %s", err))
	}

}

func usage() {
	log.Printf("Usage: subscribe [-s server] <channel> <format> \n")
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}

// TraceHandler is used to trace the message handling operation.
type handler struct {
	logger mflog.Logger
	format string
}

func (h *handler) Handle(msg *messaging.Message) error {
	switch h.format {
	case "string":
		h.logger.Info(string(msg.Payload))
	case "hex":
		h.logger.Info(hex.EncodeToString(msg.Payload))
	default:
		return errors.New("invalid format")
	}
	return nil
}

func (h *handler) Cancel() error {
	return nil
}
