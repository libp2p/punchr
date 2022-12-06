![Punchr Logo](docs/punchr-logo.jpg)

# Punchr

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> **Note**
>
> We are running a hole punching month during December 2022.
> Download the clients here (direct download links):
> 
> MacOS: [`M1/M2`](https://github.com/dennis-tra/punchr/releases/download/v0.9.0/punchr-gui-darwin-arm64.dmg) [`Intel`](https://github.com/dennis-tra/punchr/releases/download/v0.9.0/punchr-gui-darwin-amd64.dmg) (if prompted, please allow incoming connections)
> 
> Linux: [`amd64`](https://github.com/dennis-tra/punchr/releases/download/v0.9.0/punchr-gui-linux-amd64.tar.xz) [`386`](https://github.com/dennis-tra/punchr/releases/download/v0.9.0/punchr-gui-linux-386.tar.xz) | `tar -xf punchr-gui-linux-amd64.tar.xz && make user-install`
> 
> You **do not** need a running IPFS node. A CLI version and more options are available on the [release page](https://github.com/dennis-tra/punchr/releases/v0.9.0).
> 
> Optionally, [**register here**](https://forms.gle/h1ABCpS87jYmg9a48), request a personal API-Key, and receive a custom analysis of the data that you contributed.

This repo contains components to measure the [Direct Connection Upgrade through Relay (DCUtR)](https://github.com/libp2p/specs/blob/master/relay/DCUtR.md) performance.

Specifically, this repo contains the following [components](#components):

1. A [honeypot](#honeypot) that tries to attract DCUtR capable peers behind NATs.
2. A [gRPC server](#server) that exposes found peers and tracks hole punching results.
3. A [hole punching client](#clients) that fetches DCUtR capable peers from the API server, performs a hole punch to them and reports the result back.

**Dashboards:**

- [`Health Dashboard`](https://punchr.dtrautwein.eu/grafana/d/43l1QaC7z/punchr-health)
- [`Performance Dashboard`](https://punchr.dtrautwein.eu/grafana/d/F8qg0DP7k/punchr-performance)

**Talks**

IPFS þing Jul 2022         |  IPFS Camp Oct 2022 
:-------------------------:|:-------------------------:
[![IPFS þing 2022 - libp2p NAT Hole Punching Success Rate - Dennis Trautwein](https://img.youtube.com/vi/fyhZWlDbcyM/0.jpg)](https://www.youtube.com/watch?v=fyhZWlDbcyM)  |  [![IPFS Camp 2022 - Decentralized NAT Hole-Punching - Dennis Trautwein](https://img.youtube.com/vi/bzL7Y1wYth8/0.jpg)](https://www.youtube.com/watch?v=bzL7Y1wYth8)

# Table of Contents <!-- omit in toc -->
- [Punchr](#punchr)
- [Background](#background)
- [Installation](#installation)
  - [Go Client](#go-client)
    - [MacOS](#macos)
    - [Linux](#linux)
    - [Self Compilation](#self-compilation)
  - [Rust Client](#rust-client)
    - [Self Compilation](#self-compilation)
    - [Arch Linux](#arch-linux)
- [Components](#components)
  - [`honeypot`](#honeypot)
  - [`server`](#server)
  - [`go-client`](#go-client)
  - [`rust-client`](#rust-client)
- [Development](#development)
- [Deployment](#deployment)
  - [Clients](#clients)
    - [RaspberryPi](#raspberrypi)
    - [NixOS](#nixos)
- [Server](#server-1)
- [Honeypot](#honeypot-1)
- [Release](#release)
  - [Package GUI](#package-gui)
  - [go-client](#go-client-1)
- [Outcomes](#outcomes)
  - [Hole Punch Outcomes](#hole-punch-outcomes)
  - [Hole Punch Attempt Outcomes](#hole-punch-attempt-outcomes)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

# Background

![punchr flow diagram](./docs/punchr.drawio.png)

The goal is to measure the hole punching success rate. For that, we are using a **honeypot** to attract inbound connections from DCUtR capable peers behind NATs. These are then saved into a database which get served to hole punching **clients** via a **server** component. The hole punching clients ask the server if it knows about DCUtR capable peers. If it does, the clients connect to the remote peer via a relay and waits for the remote to initiate a hole punch. The result is reported back to the server.

# Installation

## Go Client

### MacOS

For macOS there is a menu bar application that you can download here: [`M1/M2`](https://github.com/dennis-tra/punchr/releases/download/v0.9.0/punchr-gui-darwin-arm64.dmg) [`Intel`](https://github.com/dennis-tra/punchr/releases/download/v0.9.0/punchr-gui-darwin-amd64.dmg)

For the CLI version head over to the [GitHub releases page](https://github.com/dennis-tra/punchr/releases) and download the appropriate binary.

### Linux

For Linux there's also a system tray application that you can download here: [`amd64`](https://github.com/dennis-tra/punchr/releases/download/v0.9.0/punchr-gui-linux-amd64.tar.xz) [`386`](https://github.com/dennis-tra/punchr/releases/download/v0.9.0/punchr-gui-linux-386.tar.xz)

You can install it by running: `tar -xf punchr-gui-linux-amd64.tar.xz && make user-install`

For the CLI version head over to the [GitHub releases page](https://github.com/dennis-tra/punchr/releases) and download the appropriate binary.

### Self Compilation

Run `make build` and find the executables in the `dist` folder. To participate in the measurement campaign you only need to pay attention to the `punchrclient` binary.

## Rust Client

### Self Compilation

```
cd rust-client

cargo build --release

export API_KEY=<YOUR_API_KEY>
./target/release/rust-client
```

```
# As an alternative via cargo
cargo run --release
```

### Arch Linux

If you are running Arch (or Manjaro), you can grab `rust-punchr` from AUR: https://aur.archlinux.org/packages/rust-punchr-bin

# Components

## `honeypot`

The honeypot operates as a DHT server and periodically walks the complete DHT to announce itself to the network. The idea is that other peers add the honeypot to their routing table. This increases the chances of peers behind NATs passing by the honeypot when they request information from the DHT network.

When the honeypot registers an inbound connection it waits until the `identify` protocol has finished and saves the following information about the remote peer to the database: PeerID, agent version, supported protocols, listen multi addresses.

<details>
   <summary>Help output:</summary>

```text
NAME:
   honeypot - A libp2p host allowing unlimited inbound connections.

USAGE:
   honeypot [global options] command [command options] [arguments...]

VERSION:
   0.6.0

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
   --crawler-count value   The number of parallel crawlers (default: 10) [$PUNCHR_HONEYPOT_CRAWLER_COUNT]
   --help, -h              show help (default: false)
   --version, -v           print the version (default: false)
```
</details>

## `server`

The server exposes a gRPC api that allows clients to query for recently seen NAT'ed DCUtR capable peers that can be probed and then report the result of the hole punching process back.

<details>
   <summary>Help output:</summary>

```text
NAME:
   punchrserver - A gRPC server that exposes peers to hole punch and tracks the results.

USAGE:
   punchrserver [global options] command [command options] [arguments...]

VERSION:
   0.9.0

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
</details>

## `go-client`

The client announces itself to the server and then periodically queries the server for peers to hole punch. If the server returns address information the client connects to the remote peer via the relay and waits for the remote to initiate a hole punch. Finally, the outcome gets reported back to the server.

Resource requirements:

- `Storage` - `~35MB`
- `Memory` - `~100MB`
- `CPU` - `~2.5%`
- `Bandwidth` - `~0.25MB`/min


<details>
<summary>Help output:</summary>

```text
NAME:
   punchrclient - A libp2p host that is capable of DCUtR.

USAGE:
   punchrclient [global options] command [command options] [arguments...]

VERSION:
   0.7.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --telemetry-host value                               To which network address should the telemetry (prometheus, pprof) server bind (default: localhost) [$PUNCHR_CLIENT_TELEMETRY_HOST]
   --telemetry-port value                               On which port should the telemetry (prometheus, pprof) server listen (default: 12001) [$PUNCHR_CLIENT_TELEMETRY_PORT]
   --server-host value                                  Where does the the punchr server listen (default: punchr.dtrautwein.eu) [$PUNCHR_CLIENT_SERVER_HOST]
   --server-port value                                  On which port listens the punchr server (default: 443) [$PUNCHR_CLIENT_SERVER_PORT]
   --server-ssl                                         Whether or not to use a SSL connection to the server. (default: true) [$PUNCHR_CLIENT_SERVER_SSL]
   --server-ssl-skip-verify                             Whether or not to skip SSL certificate verification. (default: false) [$PUNCHR_CLIENT_SERVER_SSL_SKIP_VERIFY]
   --host-count value                                   How many libp2p hosts should be used to hole punch (default: 10) [$PUNCHR_CLIENT_HOST_COUNT]
   --api-key value                                      The key to authenticate against the API [$PUNCHR_CLIENT_API_KEY]
   --key-file value                                     File where punchr saves the host identities. (default: punchrclient.keys) [$PUNCHR_CLIENT_KEY_FILE]
   --bootstrap-peers value [ --bootstrap-peers value ]  Comma separated list of multi addresses of bootstrap peers [$PUNCHR_BOOTSTRAP_PEERS]
   --disable-router-check                               Set this flag if you don't want punchr to check your router home page (default: false)
   --help, -h                                           show help (default: false)
   --version, -v                                        print the version (default: false)
```
</details>

## `rust-client`

Rust implementation of the punchr client.

Resource requirements:

- `Storage` - `~12MB`
- `Memory` - `~20MB`
- `CPU` - `~0.2%`


<details>
<summary>Help output:</summary>

```
Rust Punchr Client 0.1.0

USAGE:
    rust-client [OPTIONS]

OPTIONS:
    -h, --help                         Print help information
        --pem <PATH_TO_PEM_FILE>       Path to PEM encoded CA certificate against which the server's
                                       TLS certificate is verified [default: hardcoded CA
                                       certificate for punchr.dtrautwein.eu]
        --rounds <NUMBER_OF_ROUNDS>    Only run a fixed number of rounds
        --seed <SECRET_KEY_SEED>       Fixed value to generate a deterministic peer id
        --server <SERVER_URL>          URL and port of the punchr server. Note that the scheme ist
                                       required [default: https://punchr.dtrautwein.eu:443]
    -V, --version                      Print version information

Note: The api key for authentication is read from env value "API_KEY".
```
</details>

# Development

Run `make tools` to install all necessary tools for code generation (protobuf and database models). Specifically, this will run:

```shell
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.14.1
go install github.com/volatiletech/sqlboiler/v4@v4.6.0
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.6.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install fyne.io/fyne/v2/cmd/fyne@latest
go install github.com/fyne-io/fyne-cross@latest
```

Then start the database with `make database` or run:

```shell
docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_USER=punchr -e POSTGRES_DB=punchr postgres:14
```

Database migrations are applied automatically when starting either the honeypot or server component. To run them manually you have `make migrate-up`, `make migrate-down` and `make database-reset` at your disposal.

To create and apply a new database migration run:

```shell
migrate create -ext sql -dir migrations -seq create_some_table
make migrate-up
```


# Deployment

## Clients

### RaspberryPi

Download a `linux_armv6` or `linux_armv7` release from the [GitHub releases page](https://github.com/dennis-tra/punchr/releases), rename it to `punchrclient`, and give it execute permissions `chmod +x punchrclient`. Then you could install a systemd service at `/etc/systemd/system/punchr-client.service`:

```text
[Unit]
Description=Punchr Client
After=network-online.target

[Service]
User=pi
WorkingDirectory=/home/pi
ExecStart=/home/pi/punchrclient --api-key <some-api-key>
Restart=on-failure

[Install]
WantedBy=multiuser.target
```

To start the service run:

```shell
sudo service punchr-client start
```

### NixOS

If you're running NixOS, you can use the client option with the NixOS Module
included in the flake.

Usage:
```nix
{
  inputs.punchr = {
    url = "github:dennis-tra/punchr";
  };
   # ...

  outputs = inputs@{ self , nixpkgs }: {
      # ...Inside NixOS config
      {
         imports = [ inputs.punchr.nixosModules.client ];
         services.punchr-client.apiKey = "<API-KEY>";

         # Make sure this is readable/writable by the `punchr` user or `punchr` group.
         services.punchr-client.clientKeyFile = "/var/lib/punchr/client.keys"; # default value
      }
  };
}
```

You can run the client by itself with `nix run github:dennis-tra/punchr#client`.

# Server

Systemd service example at `/etc/systemd/system/punchr-server.service`:

```text
[Unit]
Description=Punchr Server
After=network.target

[Service]
User=punchr # your user name
WorkingDirectory=/home/punchr # configure working directory
ExecStart=/home/punchr/punchrserver # path to binary
Restart=on-failure

[Install]
WantedBy=multiuser.target
```

To start the service run:

```shell
sudo service punchr-client start
```

# Honeypot

Systemd service example at `/etc/systemd/system/punchr-honeypot.service`:

```text
[Unit]
Description=Punchr Honeypot
After=network.target

[Service]
User=punchr # your user
WorkingDirectory=/home/punchr # configure working directory
ExecStart=/home/punchr/punchrhoneypot # path to binary
Restart=on-failure

[Install]
WantedBy=multiuser.target
```

To start the service run:

```shell
sudo service punchr-honeypot start
```

# Release

## Package GUI

Run:

```shell
SIGNING_CERTIFICATE="Developer ID Application: Firstname Lastname (DevID)" NOTARIZATION_PROFILE="APPSTORE_CONNECT" make gui
```

# Outcomes

## Hole Punch Outcomes

1. `UNKNOWN` - There was no information why and how the hole punch completed
2. `NO_CONNECTION` - The client could not connect to the remote peer via any of the provided multi addresses. At the moment this is just a single relay multi address.
3. `NO_STREAM` - The client could connect to the remote peer via any of the provided multi addresses but no `/libp2p/dcutr` stream was opened within 15s. That stream is necessary to perform the hole punch.
4. `CONNECTION_REVERSED` - The client only used one or more relay multi addresses to connect to the remote peer, the `/libp2p/dcutr` stream was not opened within 15s, and we still end up with a direct connection. This means the remote peer succesfully reversed it.
5. `CANCELLED` - The user stopped the client (also returned by the rust client for quic multi addresses)
6. `FAILED` - The hole punch was attempted multiple times but none succeeded OR the `/libp2p/dcutr` was opened but we have not received the internal start event OR there was a general protocol error.
7. `SUCCESS` - Any of the three hole punch attempts succeeded.

## Hole Punch Attempt Outcomes

Any connection to a remote peer can consist of multiple attempts to hole punch a direct connection. Each individual attempt could yield the following outcomes:

1. `UNKNWON` - There was no information why and how the hole punch attempt completed
2. `DIRECT_DIAL` - The connection reversal from our side succeeded (should never happen)
3. `PROTOCOL_ERROR` - This can happen if e.g., the stream was reset mid-flight
4. `CANCELLED` - The user stopped the client
5. `TIMEOUT` - We waited for the internal start event for 15s but timed out
6. `FAILED` - We exchanged `CONNECT` and `SYNC` messages on the `/libp2p/dcutr` stream but the final direct connection attempt failed -> the hole punch was unsuccessful
7. `SUCCESS` - We were able to directly connect to the remote peer.


# Maintainers

[@dennis-tra](https://github.com/dennis-tra), 
[@elenaf9](https://github.com/elenaf9), 
[@mxinden](https://github.com/mxinden)

# Contributing

Feel free to dive in! [Open an issue](https://github.com/RichardLitt/standard-readme/issues/new) or submit PRs.

Standard Readme follows the [Contributor Covenant](http://contributor-covenant.org/version/1/3/0/) Code of Conduct.

# License

[Apache License Version 2.0](LICENSE)
