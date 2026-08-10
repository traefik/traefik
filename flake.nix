{
  description = "Dev shell for dev environment";

  inputs = {
    # Main nixpkgs (used for gnused)
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";

    # Pinned nixpkgs for kubernetes-controller-tools
    # Search: https://www.nixhub.io/packages/kubernetes-controller-tools
    nixpkgs-kct.url = "github:NixOS/nixpkgs/ee09932cedcef15aaf476f9343d1dea2cb77e261";

    # Pinned nixpkgs for golangci-lint
    # Search: https://www.nixhub.io/packages/golangci-lint
    nixpkgs-golangci.url = "github:NixOS/nixpkgs/80d901ec0377e19ac3f7bb8c035201e2e098cc97";
  };

  outputs =
    {
      nixpkgs,
      nixpkgs-kct,
      nixpkgs-golangci,
      ...
    }:
    let
      inherit (nixpkgs.lib) genAttrs;
      forEachSystem = genAttrs nixpkgs.lib.systems.flakeExposed;

      pkgsForEach = nixpkgs.legacyPackages;
      pkgsKctForEach = nixpkgs-kct.legacyPackages;
      pkgsGolangCiForEach = nixpkgs-golangci.legacyPackages;
    in
    {
      devShells = forEachSystem (system: {
        default = pkgsForEach.${system}.mkShell {
          packages = [
            pkgsForEach.${system}.gnused
            pkgsKctForEach.${system}.kubernetes-controller-tools
            pkgsGolangCiForEach.${system}.golangci-lint
          ];
        };
      });

      formatter = forEachSystem (system: pkgsForEach.${system}.nixfmt);
    };
}
