package client

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

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
	if addr == ":" {
		return
	}
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
