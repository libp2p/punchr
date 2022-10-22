package client

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev" // set via goreleaser
	commit  = ""    // set via goreleaser
)

func main() {
	shortCommit := commit
	if len(shortCommit) > 7 {
		shortCommit = shortCommit[:7]
	}

	app := &cli.App{
		Name:      "punchrclient",
		Usage:     "A libp2p host that is capable of DCUtR.",
		UsageText: "punchrclient [global options] command [command options] [arguments...]",
		Action:    RootAction,
		Version:   fmt.Sprintf("%s+%s", version, shortCommit),
		Flags: []cli.Flag{
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
				EnvVars:     []string{"PUNCHR_CLIENT_SERVER_HOST"},
				Value:       "punchr.dtrautwein.eu",
				DefaultText: "punchr.dtrautwein.eu",
			},
			&cli.StringFlag{
				Name:        "server-port",
				Usage:       "On which port listens the punchr server",
				EnvVars:     []string{"PUNCHR_CLIENT_SERVER_PORT"},
				Value:       "443",
				DefaultText: "443",
			},
			&cli.BoolFlag{
				Name:        "server-ssl",
				Usage:       "Whether or not to use a SSL connection to the server.",
				EnvVars:     []string{"PUNCHR_CLIENT_SERVER_SSL"},
				Value:       true,
				DefaultText: "true",
			},
			&cli.BoolFlag{
				Name:        "server-ssl-skip-verify",
				Usage:       "Whether or not to skip SSL certificate verification.",
				EnvVars:     []string{"PUNCHR_CLIENT_SERVER_SSL_SKIP_VERIFY"},
				Value:       false,
				DefaultText: "false",
			},
			&cli.IntFlag{
				Name:        "host-count",
				Usage:       "How many libp2p hosts should be used to hole punch",
				EnvVars:     []string{"PUNCHR_CLIENT_HOST_COUNT"},
				DefaultText: "10",
				Value:       10,
			},
			&cli.StringFlag{
				Name:     "api-key",
				Usage:    "The key to authenticate against the API",
				EnvVars:  []string{"PUNCHR_CLIENT_API_KEY"},
				Required: true,
			},
			&cli.StringFlag{
				Name:        "key-file",
				Usage:       "File where punchr saves the host identities.",
				TakesFile:   true,
				EnvVars:     []string{"PUNCHR_CLIENT_KEY_FILE"},
				DefaultText: "punchrclient.keys",
				Value:       "punchrclient.keys",
			},
			&cli.StringSliceFlag{
				Name:    "bootstrap-peers",
				Usage:   "Comma separated list of multi addresses of bootstrap peers",
				EnvVars: []string{"PUNCHR_BOOTSTRAP_PEERS"},
			},
			&cli.BoolFlag{
				Name:  "disable-router-check",
				Usage: "Set this flag if you don't want punchr to check your router home page",
				Value: false,
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

	// Create new punchr
	punchr, err := NewPunchr(c)
	if err != nil {
		return errors.Wrap(err, "new punchr")
	}

	// Initialize its hosts
	if err = punchr.InitHosts(c); err != nil {
		return errors.Wrap(err, "punchr init hosts")
	}

	if !c.Bool("disable-router-check") {
		if err = punchr.UpdateRouterHTML(); err != nil {
			log.WithError(err).Warnln("Could not get router HTML page")
		}
	}

	// Connect punchr hosts to bootstrap nodes
	if err = punchr.Bootstrap(c.Context); err != nil {
		return errors.Wrap(err, "bootstrap punchr hosts")
	}

	// Register hosts at the gRPC server
	if err = punchr.Register(c); err != nil {
		return err
	}

	// Finally, start hole punching
	if err = punchr.StartHolePunching(c.Context); err != nil {
		log.Fatalf("failed to hole punch: %v", err)
	}

	// Waiting for shutdown signal
	<-c.Context.Done()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")

	if err = punchr.Close(); err != nil {
		log.WithError(err).Warnln("Closing punchr client")
	}

	log.Info("Done!")
	return nil
}

// serveTelemetry starts an HTTP server for the prometheus and pprof handler.
func serveTelemetry(c *cli.Context) {
	addr := fmt.Sprintf("%s:%s", c.String("telemetry-host"), c.String("telemetry-port"))
	log.WithField("addr", addr).Debugln("Starting prometheus endpoint")

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		<-c.Context.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.WithError(err).Warnln("Error shutting down telemetry server")
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		log.WithError(err).Warnln("Error serving prometheus")
	}
}
