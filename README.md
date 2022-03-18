# Punchr

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

This repo contains components to measure [Direct Connection Upgrade through Relay (DCUtR)](https://github.com/libp2p/specs/blob/master/relay/DCUtR.md) performance.

Specifically, this repo contains:

1. A honeypot that tries to attract DCUtR capable peers behind NATs.
2. A gRPC server that exposes found peers and tracks hole punching results.
3. A hole punching client that fetches DCUtR capable peers from the API server, performs a hole punch to them and reports the result back.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Background](#background)
- [Install](#install)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Background

![punchr flow diagram](./docs/punchr.drawio.png)

The goal is to measure the hole punching success rate. For that, we are using a **honeypot** to attract inbound connections from DCUtR capable peers behind NATs. These are then saved into a database which get served to hole punching **clients** via a **server** component. The hole punching clients ask the server if it knows about DCUtR capable peers. If it does the clients connect to the remote peer via a relay and waits for the remote to initiate a hole punch. The result is reported back to the server.

## Components

### `honeypot`

The honeypot operates as a DHT server and periodically walks the complete DHT to announce itself to the network. The idea is that other peers add the honeypot to their routing table. This increases the chances of peers behind NATs passing by the honeypot when they request information from the DHT network.

When the honeypot registers an inbound connection it waits until the `identify` protocol has finished and saves the following information about the remote peer to the database: PeerID, agent version, supported protocols, listen multi addresses.

Help output:

```
NAME:
   honeypot - A libp2p host allowing unlimited inbound connections.

USAGE:
   honeypot [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value            On which port should the libp2p host listen (default: 11000) [$PUNCHR_HONEYPOT_PORT]
   --telemetry-host value  To which network address should the telemetry (prometheus, pprof) server bind (default: localhost) [$PUNCHR_HONEYPOT_TELEMETRY_HOST]
   --telemetry-port value  On which port should the telemetry (prometheus, pprof) server listen (default: 11001) [$PUNCHR_HONEYPOT_TELEMETRY_PORT]
   --db-host value         On which host address can the database be reached (default: localhost) [$PUNCHR_HONEYPOT_DATABASE_HOST]
   --db-port value         On which port can the database be reached (default: 5432) [$PUNCHR_HONEYPOT_DATABASE_PORT]
   --db-name value         The name of the database to use (default: punchr) [$PUNCHR_HONEYPOT_DATABASE_NAME]
   --db-password value     The password for the database to use (default: password) [$PUNCHR_HONEYPOT_DATABASE_PASSWORD]
   --db-user value         The user with which to access the database to use (default: punchr) [$PUNCHR_HONEYPOT_DATABASE_USER]
   --db-sslmode value      The sslmode to use when connecting the the database (default: disable) [$PUNCHR_HONEYPOT_DATABASE_SSL_MODE]
   --key FILE              Load private key for peer ID from FILE (default: honeypot.key) [$PUNCHR_HONEYPOT_KEY_FILE]
   --help, -h              show help (default: false)
   --version, -v           print the version (default: false)
```
### `server`

The server exposes a gRPC api that allows clients to query for recently seen NAT'ed DCUtR capable peers that can be probed and then report the result of the hole punching process back.

Help output:

```shell
NAME:
   punchrserver - A gRPC server that exposes peers to hole punch and tracks the results.

USAGE:
   punchrserver [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value            On which port should the gRPC host listen (default: 10000) [$PUNCHR_SERVER_PORT]
   --telemetry-host value  To which network address should the telemetry (prometheus, pprof) server bind (default: localhost) [$PUNCHR_SERVER_TELEMETRY_HOST]
   --telemetry-port value  On which port should the telemetry (prometheus, pprof) server listen (default: 10001) [$PUNCHR_SERVER_TELEMETRY_PORT]
   --db-host value         On which host address can the database be reached (default: localhost) [$PUNCHR_SERVER_DATABASE_HOST]
   --db-port value         On which port can the database be reached (default: 5432) [$PUNCHR_SERVER_DATABASE_PORT]
   --db-name value         The name of the database to use (default: punchr) [$PUNCHR_SERVER_DATABASE_NAME]
   --db-password value     The password for the database to use (default: password) [$PUNCHR_SERVER_DATABASE_PASSWORD]
   --db-user value         The user with which to access the database to use (default: punchr) [$PUNCHR_SERVER_DATABASE_USER]
   --db-sslmode value      The sslmode to use when connecting the the database (default: disable) [$PUNCHR_SERVER_DATABASE_SSL_MODE]
   --help, -h              show help (default: false)
   --version, -v           print the version (default: false)
```

### `client`

The client announces itself to the server and then periodically queries the server for peers to hole punch. If the server returns address information the client connects to the remote peer via the relay and waits for the remote to initiate a hole punch. Finally, the outcome gets reported back to the server.

Help output:
```
NAME:
   punchrclient - A libp2p host that is capable of DCUtR.

USAGE:
   punchrclient [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value            On which port should the libp2p host listen (default: 12000) [$PUNCHR_CLIENT_PORT]
   --telemetry-host value  To which network address should the telemetry (prometheus, pprof) server bind (default: localhost) [$PUNCHR_CLIENT_TELEMETRY_HOST]
   --telemetry-port value  On which port should the telemetry (prometheus, pprof) server listen (default: 12001) [$PUNCHR_CLIENT_TELEMETRY_PORT]
   --server-host value     Where does the the punchr server listen (default: localhost) [$PUNCHR_CLIENT_TELEMETRY_HOST]
   --server-port value     On which port listens the punchr server (default: 10000) [$PUNCHR_CLIENT_TELEMETRY_PORT]
   --key FILE              Load private key for peer ID from FILE (default: punchr.key) [$PUNCHR_CLIENT_KEY_FILE]
   --help, -h              show help (default: false)
   --version, -v           print the version (default: false)
```

## Install

Run `make build` and find the executables in the `dist` folder.

The honeypot listens on port `10000`, the server on port `11000` and clients on `12000`. All components expose prometheus and pprof telemetry on `10001`, `11001`, and `12001` respectively.

## Development

Run `make tools` to install all necessary tools for code generation (protobuf and database models). Specifically, this will run:

```shell
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1
go install github.com/volatiletech/sqlboiler/v4@v4.6.0
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.6.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Then start the database with `make database` or run:

```shell
docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=punchr -e POSTGRES_DB=punchr postgres:13
```

Database migrations are applied automatically when starting either the honeypot or server component. To run them manually you have `make migrate-up`, `make migrate-down` and `make database-reset` at your disposal.

To create and apply a new database migration run:

```shell
migrate create -ext sql -dir migrations -seq create_some_table
make migrate-up
```

## Maintainers

[@dennis-tra](https://github.com/dennis-tra).

## Contributing

Feel free to dive in! [Open an issue](https://github.com/RichardLitt/standard-readme/issues/new) or submit PRs.

Standard Readme follows the [Contributor Covenant](http://contributor-covenant.org/version/1/3/0/) Code of Conduct.

## License

[MIT](LICENSE)
