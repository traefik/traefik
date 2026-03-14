{
  description = "Dev shell for dev environment";

  inputs = {
    # Main nixpkgs (used for gnused)
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    # Pinned nixpkgs for kubernetes-controller-tools
    # Search: https://www.nixhub.io/packages/kubernetes-controller-tools
    nixpkgs-kct.url = "github:NixOS/nixpkgs/ee09932cedcef15aaf476f9343d1dea2cb77e261";

    # Pinned nixpkgs for golangci-lint
    # Search: https://www.nixhub.io/packages/golangci-lint
    nixpkgs-golangci.url = "github:NixOS/nixpkgs/80d901ec0377e19ac3f7bb8c035201e2e098cc97";

    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, nixpkgs-kct, nixpkgs-golangci, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        pkgs-kct = import nixpkgs-kct {
          inherit system;
        };

        pkgs-golangci = import nixpkgs-golangci {
          inherit system;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs-kct.kubernetes-controller-tools
            pkgs.gnused
            pkgs-golangci.golangci-lint
          ];
        };
      }
    );
}