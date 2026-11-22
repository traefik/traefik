{
  description = "Dev shell for dev environment";

  inputs = {
    # Main nixpkgs (used for gnused)
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";

    # Pinned nixpkgs for kubernetes-controller-tools (0.21.0)
    # Search: https://www.nixhub.io/packages/kubernetes-controller-tools
    nixpkgs-kct.url = "github:NixOS/nixpkgs/c27cdad491a991b11ed731760aa2ef8db0cb0410";

    # Pinned nixpkgs for golangci-lint (2.14.0)
    # Search: https://www.nixhub.io/packages/golangci-lint
    nixpkgs-golangci.url = "github:NixOS/nixpkgs/55ba7f49ef2962b42cbd126522b7df5f95037679";

    # Pinned nixpkgs for shellcheck (0.11.0)
    # Search: https://www.nixhub.io/packages/shellcheck
    nixpkgs-shellcheck.url = "github:NixOS/nixpkgs/55ba7f49ef2962b42cbd126522b7df5f95037679";
  };

  outputs =
    {
      nixpkgs,
      nixpkgs-kct,
      nixpkgs-golangci,
      nixpkgs-shellcheck,
      ...
    }:
    let
      inherit (nixpkgs.lib) genAttrs attrNames;
      forEachSystem = genAttrs nixpkgs.lib.systems.flakeExposed;

      pkgsForEach = nixpkgs.legacyPackages;
      pkgsKctForEach = nixpkgs-kct.legacyPackages;
      pkgsGolangCiForEach = nixpkgs-golangci.legacyPackages;
      pkgsShellcheckForEach = nixpkgs-shellcheck.legacyPackages;

      # The CI uses the latest Go patch release (setup-go with check-latest),
      # which nixpkgs does not always provide yet, so the official binary is used.
      # Hashes: https://go.dev/dl/?mode=json&include=all
      goVersion = "1.27.3";
      goBinaries = {
        x86_64-linux = {
          platform = "linux-amd64";
          sha256 = "65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40";
        };
        aarch64-linux = {
          platform = "linux-arm64";
          sha256 = "a06e99309dbcbdca1a8ccd2282c3ee146e55930886e26b4f1193269587165f49";
        };
        x86_64-darwin = {
          platform = "darwin-amd64";
          sha256 = "b6f8e931e4ab497fcde6da8ae4499a58f9d258e55c8a80824a516eae5950a750";
        };
        aarch64-darwin = {
          platform = "darwin-arm64";
          sha256 = "dd3908cf20959d798c1e06ebe5b72213f0505ce01373b40ce3110975a6cd99c2";
        };
      };

      goForEach =
        system:
        let
          pkgs = pkgsForEach.${system};
          bin = goBinaries.${system};
        in
        pkgs.stdenvNoCC.mkDerivation {
          pname = "go-bin";
          version = goVersion;

          src = pkgs.fetchurl {
            url = "https://go.dev/dl/go${goVersion}.${bin.platform}.tar.gz";
            inherit (bin) sha256;
          };

          # The official binaries are used as is, fixup would alter them.
          dontConfigure = true;
          dontBuild = true;
          dontFixup = true;

          installPhase = ''
            runHook preInstall
            mkdir -p $out/share $out/bin
            cp -r . $out/share/go
            ln -s $out/share/go/bin/go $out/share/go/bin/gofmt $out/bin/
            runHook postInstall
          '';
        };
    in
    {
      devShells = genAttrs (attrNames goBinaries) (system: {
        default = pkgsForEach.${system}.mkShell {
          packages = [
            pkgsForEach.${system}.gnused
            (goForEach system)
            pkgsKctForEach.${system}.kubernetes-controller-tools
            pkgsGolangCiForEach.${system}.golangci-lint
            pkgsShellcheckForEach.${system}.shellcheck
          ];

          # Matches the CI workflows (build, integration and conformance tests).
          CGO_ENABLED = "0";
        };
      });

      formatter = forEachSystem (system: pkgsForEach.${system}.nixfmt);
    };
}
