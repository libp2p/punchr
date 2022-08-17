{
  description = "Punchr";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/release-22.05";

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          system = system;
        };
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
          vendorSha256 = "sha256-nb0oq3vA8ANlppmpKzL61n6eS96eXiBI6Mumxu00rII=";

          meta = with pkgs.lib; {
            description = "";
            homepage = "https://github.com/dennis-tra/punchr";
            license = licenses.mit;
            maintainers = with maintainers; [ "marcopolo" ];
            platforms = platforms.linux ++ platforms.darwin;
          };
        };
        defaultPackage = self.packages.${system}.client;
        devShell = pkgs.mkShell {
          buildInputs = [ pkgs.go ];
        };
      })) // {
      nixosModules.client = { config, pkgs, ... }:
        let cfg = config.services.punchr-client;
        in
        {
          options.services.punchr-client = {
            apiKey = nixpkgs.lib.mkOption {
              description = "Punchr API Key";
              type = nixpkgs.lib.types.str;
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
            systemd.services.punchr-client = {
              description = "punchr-client";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" ];
              serviceConfig = {
                Environment = "PUNCHR_CLIENT_API_KEY=${cfg.apiKey} PUNCHR_CLIENT_KEY_FILE=${cfg.clientKeyFile} NEBULA_BOOTSTRAP_PEERS=${cfg.bootstrapPeers}";
                ExecStart = "${
                  self.packages.${pkgs.system}.client
                  }/bin/client";
                Restart = "always";
                RestartSec = "1min";
                User = "punchr";
              };
            };
          };
        };
    };
}
