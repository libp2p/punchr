package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	"github.com/dennis-tra/punchr/pkg/db"
	"github.com/dennis-tra/punchr/pkg/pb"
)

func main() {
	app := &cli.App{
		Name:      "punchrserver",
		Usage:     "A gRPC server that exposes peers to hole punch and tracks the results.",
		UsageText: "punchrserver [global options] command [command options] [arguments...]",
		Action:    RootAction,
		Version:   "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Usage:       "On which port should the gRPC host listen",
				EnvVars:     []string{"PUNCHR_SERVER_PORT"},
				Value:       "10000",
				DefaultText: "10000",
			},
			&cli.StringFlag{
				Name:        "telemetry-host",
				Usage:       "To which network address should the telemetry (prometheus, pprof) server bind",
				EnvVars:     []string{"PUNCHR_SERVER_TELEMETRY_HOST"},
				Value:       "localhost",
				DefaultText: "localhost",
			},
			&cli.StringFlag{
				Name:        "telemetry-port",
				Usage:       "On which port should the telemetry (prometheus, pprof) server listen",
				EnvVars:     []string{"PUNCHR_SERVER_TELEMETRY_PORT"},
				Value:       "10001",
				DefaultText: "10001",
			},
			&cli.StringFlag{
				Name:        "db-host",
				Usage:       "On which host address can the database be reached",
				EnvVars:     []string{"PUNCHR_SERVER_DATABASE_HOST"},
				DefaultText: "localhost",
				Value:       "localhost",
			},
			&cli.StringFlag{
				Name:        "db-port",
				Usage:       "On which port can the database be reached",
				EnvVars:     []string{"PUNCHR_SERVER_DATABASE_PORT"},
				DefaultText: "5432",
				Value:       "5432",
			},
			&cli.StringFlag{
				Name:        "db-name",
				Usage:       "The name of the database to use",
				EnvVars:     []string{"PUNCHR_SERVER_DATABASE_NAME"},
				DefaultText: "punchr",
				Value:       "punchr",
			},
			&cli.StringFlag{
				Name:        "db-password",
				Usage:       "The password for the database to use",
				EnvVars:     []string{"PUNCHR_SERVER_DATABASE_PASSWORD"},
				DefaultText: "password",
				Value:       "password",
			},
			&cli.StringFlag{
				Name:        "db-user",
				Usage:       "The user with which to access the database to use",
				EnvVars:     []string{"PUNCHR_SERVER_DATABASE_USER"},
				DefaultText: "punchr",
				Value:       "punchr",
			},
			&cli.StringFlag{
				Name:        "db-sslmode",
				Usage:       "The sslmode to use when connecting the the database",
				EnvVars:     []string{"PUNCHR_SERVER_DATABASE_SSL_MODE"},
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

	// Initialize gRPC server
	s, lis, err := initGrpcServer(c, err)
	if err != nil {
		return err
	}
	pb.RegisterPunchrServiceServer(s, &Server{DBClient: dbClient})

	// Start gRPC server
	log.WithField("addr", lis.Addr().String()).Infoln("Starting server")
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalln(errors.Wrap(err, "failed to serve"))
		}
	}()

	// Waiting for shutdown signal
	<-c.Context.Done()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")

	// Closing database connection
	if err = dbClient.Close(); err != nil {
		log.WithError(err).Warnln("closing db client")
	}

	// Stopping gRPC server
	s.Stop()

	log.Info("Done!")
	return nil
}

// serveTelemetry starts an HTTP server for the prometheus and pprof handler.
func serveTelemetry(c *cli.Context) {
	addr := fmt.Sprintf("%s:%s", c.String("telemetry-host"), c.String("telemetry-port"))
	log.WithField("addr", addr).Infoln("Starting telemetry endpoints")
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Warnln("Error serving prometheus")
	}
}

func initGrpcServer(c *cli.Context, err error) (*grpc.Server, net.Listener, error) {
	logger := log.StandardLogger()
	logger.SetLevel(log.DebugLevel)
	logEntry := log.NewEntry(logger)
	grpc_logrus.ReplaceGrpcLogger(logEntry)
	opts := []grpc_logrus.Option{
		grpc_logrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_s", duration.Seconds()
		}),
	}
	s := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_logrus.UnaryServerInterceptor(logEntry, opts...),
			grpc_prometheus.UnaryServerInterceptor,
		))
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", c.String("port")))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to listen")
	}

	return s, lis, nil
}
