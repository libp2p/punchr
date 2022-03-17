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
		Name:      "punchr",
		Usage:     "A libp2p host that is capable of DCUtR.",
		UsageText: "punchr [global options] command [command options] [arguments...]",
		Action:    RootAction,
		Version:   "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Usage:       "On which port should the libp2p host listen",
				EnvVars:     []string{"PUNCHR_PORT"},
				Value:       "13500",
				DefaultText: "13500",
			},
			&cli.StringFlag{
				Name:        "telemetry-host",
				Usage:       "To which network address should the telemetry (prometheus, pprof) server bind",
				EnvVars:     []string{"PUNCHR_TELEMETRY_HOST"},
				Value:       "localhost",
				DefaultText: "localhost",
			},
			&cli.StringFlag{
				Name:        "telemetry-port",
				Usage:       "On which port should the telemetry (prometheus, pprof) server listen",
				EnvVars:     []string{"PUNCHR_TELEMETRY_PORT"},
				Value:       "13000",
				DefaultText: "13000",
			},
			&cli.StringFlag{
				Name:        "api-host",
				Usage:       "Where does the the punchr API listen",
				EnvVars:     []string{"PUNCHR_TELEMETRY_HOST"},
				Value:       "localhost",
				DefaultText: "localhost",
			},
			&cli.StringFlag{
				Name:        "api-port",
				Usage:       "On which port listens the punchr API",
				EnvVars:     []string{"PUNCHR_TELEMETRY_PORT"},
				Value:       "12500",
				DefaultText: "12500",
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

	h, err := InitHost(c.Context, c.App.Version, c.String("port"))
	if err != nil {
		return errors.Wrap(err, "init host")
	}

	addr := fmt.Sprintf("%s:%s", c.String("api-host"), c.String("api-port"))
	conn, err := grpc.Dial(addr)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	h.APIClient = pb.NewPunchrServiceClient(conn)

	// Waiting for shutdown signal
	<-c.Context.Done()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")

	if err = conn.Close(); err != nil {
		log.WithError(err).Warnln("Closing gRPC API connection")
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