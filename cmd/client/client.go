package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/dennis-tra/punchr/pkg/pb"
)

func main() {
	app := &cli.App{
		Name:      "punchrclient",
		Usage:     "A libp2p host that is capable of DCUtR.",
		UsageText: "punchrclient [global options] command [command options] [arguments...]",
		Action:    RootAction,
		Version:   "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Usage:       "On which port should the libp2p host listen",
				EnvVars:     []string{"PUNCHR_CLIENT_PORT"},
				Value:       "12000",
				DefaultText: "12000",
			},
			&cli.StringFlag{
				Name:        "telemetry-host",
				Usage:       "To which network address should the telemetry (prometheus, pprof) server bind",
				EnvVars:     []string{"PUNCHR_CLIENT_TELEMETRY_HOST"},
				Value:       "localhost",
				DefaultText: "localhost",
			},
			&cli.StringFlag{
				Name:        "telemetry-port",
				Usage:       "On which port should the telemetry (prometheus, pprof) server listen",
				EnvVars:     []string{"PUNCHR_CLIENT_TELEMETRY_PORT"},
				Value:       "12001",
				DefaultText: "12001",
			},
			&cli.StringFlag{
				Name:        "server-host",
				Usage:       "Where does the the punchr server listen",
				EnvVars:     []string{"PUNCHR_CLIENT_TELEMETRY_HOST"},
				Value:       "localhost",
				DefaultText: "localhost",
			},
			&cli.StringFlag{
				Name:        "server-port",
				Usage:       "On which port listens the punchr server",
				EnvVars:     []string{"PUNCHR_CLIENT_TELEMETRY_PORT"},
				Value:       "10000",
				DefaultText: "10000",
			},
			&cli.StringFlag{
				Name:        "key",
				Usage:       "Load private key for peer ID from `FILE`",
				TakesFile:   true,
				EnvVars:     []string{"PUNCHR_CLIENT_KEY_FILE"},
				DefaultText: "punchr.key",
				Value:       "punchr.key",
			},
		},
		EnableBashCompletion: true,
	}

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Errorf("error: %v\n", err)
		os.Exit(1)
	}
}

func RootAction(c *cli.Context) error {
	// Start telemetry endpoints
	go serveTelemetry(c)

	addr := fmt.Sprintf("%s:%s", c.String("server-host"), c.String("server-port"))
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	h, err := InitHost(c, c.String("port"))
	if err != nil {
		return errors.Wrap(err, "init host")
	}

	h.client = pb.NewPunchrServiceClient(conn)

	// Connect punchr host to bootstrap nodes
	if err := h.Bootstrap(c.Context); err != nil {
		return errors.Wrap(err, "bootstrap host")
	}

	// Register host at the gRPC server
	if err := h.RegisterHost(c.Context); err != nil {
		return err
	}

	// Finally, start hole punching
	if err = h.StartHolePunching(c.Context); err != nil {
		log.Fatalf("failed to hole punch: %v", err)
	}

	// Waiting for shutdown signal
	<-c.Context.Done()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")

	if err = conn.Close(); err != nil {
		log.WithError(err).Warnln("Closing gRPC server connection")
	}

	log.Info("Done!")
	return nil
}

// serveTelemetry starts an HTTP server for the prometheus and pprof handler.
func serveTelemetry(c *cli.Context) {
	addr := fmt.Sprintf("%s:%s", c.String("telemetry-host"), c.String("telemetry-port"))
	log.WithField("addr", addr).Debugln("Starting prometheus endpoint")
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Warnln("Error serving prometheus")
	}
}
