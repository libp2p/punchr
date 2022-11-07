package client

import (
	"github.com/urfave/cli/v2"
)

var Version = "dev"

var App = &cli.App{
	Name:      "punchrclient",
	Usage:     "A libp2p host that is capable of DCUtR.",
	UsageText: "punchrclient [global options] command [command options] [arguments...]",
	Action:    RootAction,
	Version:   Version,
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
			Name:    "api-key",
			Usage:   "The key to authenticate against the API. If not set, it's read from $XDG_CONFIG_HOME/punchr/api-key.txt",
			EnvVars: []string{"PUNCHR_CLIENT_API_KEY"},
		},
		&cli.StringFlag{
			Name:        "key-file",
			Usage:       "File where punchr saves the host identities.",
			TakesFile:   true,
			EnvVars:     []string{"PUNCHR_CLIENT_KEY_FILE"},
			DefaultText: "$XDG_CONFIG_HOME/punchr/client.keys",
			Value:       "$XDG_CONFIG_HOME/punchr/client.keys",
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
