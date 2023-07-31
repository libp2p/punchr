{
  description = "Punchr";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/release-22.05";
  inputs.nixpkgs-darwin.url = "github:nixos/nixpkgs/nixpkgs-22.05-darwin";
  inputs.rust-overlay.url = "github:oxalica/rust-overlay";


  outputs = { self, flake-utils, rust-overlay, ... }@inputs:
    (flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [ (import rust-overlay) ];
        nixpkgs = if system == "aarch64-darwin" then inputs.nixpkgs-darwin else inputs.nixpkgs;
        pkgs = import nixpkgs { inherit system overlays; };
        rustStable = pkgs.rust-bin.stable.latest.default.override {
          extensions = [ "rust-src" ];
        };
        rustPlatform = (pkgs.makeRustPlatform {
          cargo = pkgs.rust-bin.stable.latest.default;
          rustc = pkgs.rust-bin.stable.latest.default;
        });
      in
      {
        packages.client = pkgs.buildGo118Module rec {
          src = ./.;

          pname = "client";
          version = "0.4.0";
          subPackages = [ "cmd/client" ];
          checkPhase = "";


          # Sha of go modules. To update uncomment the below line and comment
          # out the current sha. Then update the sha.
          # vendorSha256 = pkgs.lib.fakeSha256;
          vendorSha256 = "sha256-qR1kWFY3un9TRKu/3rEMnn+shkDTkh+Lsg/Om1Klf4w=";

          meta = with pkgs.lib; {
            description = "";
            homepage = "https://github.com/dennis-tra/punchr";
            license = licenses.mit;
            maintainers = with maintainers; [ "marcopolo" ];
            platforms = platforms.linux ++ platforms.darwin;
          };
        };
        packages.rust-client = rustPlatform.buildRustPackage rec {
          src = ./rust-client;

          pname = "rust-client";
          version = "0.0.1";
          checkPhase = "";

          nativeBuildInputs = [
            pkgs.cmake
            pkgs.rustfmt
            pkgs.protobuf
          ];

          buildInputs = [
            pkgs.libiconv
          ] ++ (if system == "aarch64-darwin" then [
            pkgs.darwin.apple_sdk.frameworks.SystemConfiguration
          ] else [ ]);

          cargoLock = {
            lockFile = ./rust-client/Cargo.lock;
            outputHashes = {
              # Update the hash by using a fake hash, and then updating it with
              # the correct one when the build fails
              "libp2p-0.52.0" = pkgs.lib.fakeSha256;
              # "libp2p-0.52.0" = "sha256-IVTm13gCmMZl4bLODaKFybR5r2R2ObTs8G//GDwoH1A=";
            };
          };

          PUNCHR_PROTO = ./punchr.proto;

          meta = with pkgs.lib; {
            description = "Hole punching stats gatherer, in Rust";
            homepage = "https://github.com/dennis-tra/punchr";
            license = licenses.mit;
            maintainers = with maintainers; [ "marcopolo" ];
            platforms = platforms.linux ++ platforms.darwin;
          };
        };

        defaultPackage = self.packages.${system}.client;

        devShell = pkgs.mkShell {
          PUNCHR_PROTO = ./punchr.proto;
          buildInputs = [
            pkgs.go
            rustStable
            pkgs.rustfmt
            pkgs.libiconv
            pkgs.protobuf
          ];
        };
      })) // {
      nixosModules.client = { config, pkgs, ... }:
        let
          nixpkgs = inputs.nixpkgs;
          lib = nixpkgs.lib;
          cfg = config.services.punchr-client;
        in
        {
          options.services.punchr-client = {
            apiKey = nixpkgs.lib.mkOption {
              description = "Punchr API Key";
              type = nixpkgs.lib.types.str;
            };
            enableGoClient = nixpkgs.lib.mkOption {
              description = "Enable the Go Client";
              type = nixpkgs.lib.types.bool;
              default = true;
            };
            enableRustClient = nixpkgs.lib.mkOption {
              description = "Enable the Rust Client";
              type = nixpkgs.lib.types.bool;
              default = true;
            };
            clientKeyFile = nixpkgs.lib.mkOption {
              description = "Path to file where punchr saves the host identities. Should be writable by punchr-client";
              type = nixpkgs.lib.types.str;
              default = "/var/lib/punchr/client.keys";
              example = "/var/lib/punchr/client.keys";
            };
            bootstrapPeers = nixpkgs.lib.mkOption {
              description = "Comma separated list of multi addresses of bootstrap peers  (accepts multiple inputs)";
              type = nixpkgs.lib.types.str;
              default = "/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN,/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa,/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb,/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt,/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ,/ip4/147.75.83.83/tcp/4001/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb,/ip4/147.75.77.187/tcp/4001/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa,/ip4/147.75.109.29/tcp/4001/p2p/QmZa1sAxajnQjVM8WjWXoMbmPd7NsWhfKsPkErzpm9wGkp";
              example = "/ip4/147.75.83.83/tcp/4001/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb";
            };
          };
          config = {
            users.users.punchr.isSystemUser = true;
            users.users.punchr.group = "punchr";
            users.groups.punchr = { };
            systemd.services = {
              punchr-go-client = (lib.mkIf cfg.enableGoClient {
                description = "punchr-go-client";
                wantedBy = [ "multi-user.target" ];
                after = [ "network.target" ];
                serviceConfig = {
                  Environment = "PUNCHR_CLIENT_API_KEY=${cfg.apiKey} PUNCHR_CLIENT_KEY_FILE=${cfg.clientKeyFile} NEBULA_BOOTSTRAP_PEERS=${cfg.bootstrapPeers}";
                  ExecStart = "${self.packages.${pkgs.system}.client}/bin/client";
                  Restart = "always";
                  RestartSec = "1min";
                  User = "punchr";
                };
              });

              punchr-rust-client = (lib.mkIf cfg.enableRustClient {
                description = "punchr-rust-client";
                wantedBy = [ "multi-user.target" ];
                after = [ "network.target" ];
                serviceConfig = {
                  Environment = "API_KEY=${cfg.apiKey}";
                  ExecStart = "${self.packages.${pkgs.system}.rust-client}/bin/rust-client";
                  Restart = "always";
                  RestartSec = "1min";
                  User = "punchr";
                };
              });
            };
          };
        };
    };
}
