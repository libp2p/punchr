package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/punchr/pkg/db"
)

var (
	connOpen = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "connection_open",
			Namespace: "honeypot",
			Help:      "The connection open events of the honeypot libp2p host",
		},
		[]string{"direction"},
	)

	connClose = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "connection_close",
			Namespace: "honeypot",
			Help:      "The connection close events of the honeypot libp2p host",
		},
		[]string{"direction"},
	)

	openConns = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "open_connections_total",
			Namespace: "honeypot",
			Help:      "The currently open connection of the honeypot libp2p host",
		},
		[]string{"direction"},
	)

	handleConnDur = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:      "handle_connected_duration",
			Namespace: "honeypot",
			Help:      "The time it took to handle a connection establishment event",
			Buckets:   prometheus.DefBuckets,
		},
	)
)

func main() {
	app := &cli.App{
		Name:      "honeypot",
		Usage:     "A libp2p host allowing unlimited inbound connections.",
		UsageText: "honeypot [global options] command [command options] [arguments...]",
		Action:    RootAction,
		Version:   "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Usage:       "On which port should the libp2p host listen",
				EnvVars:     []string{"PUNCHR_HONEYPOT_PORT"},
				Value:       "10500",
				DefaultText: "10500",
			},
			&cli.StringFlag{
				Name:        "telemetry-host",
				Usage:       "To which network address should the telemetry (prometheus, pprof) server bind",
				EnvVars:     []string{"PUNCHR_HONEYPOT_TELEMETRY_HOST"},
				Value:       "localhost",
				DefaultText: "localhost",
			},
			&cli.StringFlag{
				Name:        "telemetry-port",
				Usage:       "On which port should the telemetry (prometheus, pprof) server listen",
				EnvVars:     []string{"PUNCHR_HONEYPOT_TELEMETRY_PORT"},
				Value:       "10000",
				DefaultText: "10000",
			},
			&cli.StringFlag{
				Name:        "db-host",
				Usage:       "On which host address can the database be reached",
				EnvVars:     []string{"PUNCHR_HONEYPOT_DATABASE_HOST"},
				DefaultText: "localhost",
				Value:       "localhost",
			},
			&cli.StringFlag{
				Name:        "db-port",
				Usage:       "On which port can the database be reached",
				EnvVars:     []string{"PUNCHR_HONEYPOT_DATABASE_PORT"},
				DefaultText: "5432",
				Value:       "5432",
			},
			&cli.StringFlag{
				Name:        "db-name",
				Usage:       "The name of the database to use",
				EnvVars:     []string{"PUNCHR_HONEYPOT_DATABASE_NAME"},
				DefaultText: "punchr",
				Value:       "punchr",
			},
			&cli.StringFlag{
				Name:        "db-password",
				Usage:       "The password for the database to use",
				EnvVars:     []string{"PUNCHR_HONEYPOT_DATABASE_PASSWORD"},
				DefaultText: "password",
				Value:       "password",
			},
			&cli.StringFlag{
				Name:        "db-user",
				Usage:       "The user with which to access the database to use",
				EnvVars:     []string{"PUNCHR_HONEYPOT_DATABASE_USER"},
				DefaultText: "punchr",
				Value:       "punchr",
			},
			&cli.StringFlag{
				Name:        "db-sslmode",
				Usage:       "The sslmode to use when connecting the the database",
				EnvVars:     []string{"PUNCHR_HONEYPOT_DATABASE_SSL_MODE"},
				DefaultText: "disable",
				Value:       "disable",
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

	// Initialize database connection
	dbClient, err := db.NewClient(c)
	if err != nil {
		return errors.Wrap(err, "new db client")
	}

	// Initialize honeypot libp2p host
	h, err := InitHost(c.Context, c.App.Version, c.String("port"), dbClient)
	if err != nil {
		return errors.Wrap(err, "init host")
	}

	// Connect honeypot host to bootstrap nodes
	if err := h.Bootstrap(c.Context); err != nil {
		return errors.Wrap(err, "bootstrap host")
	}

	// Slowly start passing by other libp2p hosts for them to add us to their routing table.
	go h.WalkDHT()

	// Waiting for shutdown signal
	<-c.Context.Done()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")

	if err = h.Close(); err != nil {
		log.WithError(err).Warnln("closing libp2p host")
	}

	if err = dbClient.Close(); err != nil {
		log.WithError(err).Warnln("closing db client")
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
