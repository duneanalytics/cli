{
  description = "Dune CLI: query data, manage visualizations, and build dashboards from your terminal";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      # Flakes can't see git tags, so the version comes from the commit date.
      # commit and date match what GoReleaser stamps into release builds.
      ts = self.lastModifiedDate; # YYYYMMDDHHMMSS
      part = start: len: builtins.substring start len ts;
      version = "0-unstable-${part 0 4}-${part 4 2}-${part 6 2}";
      commit = self.rev or self.dirtyRev or "unknown";
      date = "${part 0 4}-${part 4 2}-${part 6 2}T${part 8 2}:${part 10 2}:${part 12 2}Z";
    in
    {
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "dune-cli";
          inherit version;
          src = self;

          # Update when go.mod or go.sum changes: set it to lib.fakeHash, run
          # `nix build`, and copy the hash from the error. CI fails while it is stale.
          vendorHash = "sha256-7hB7LC+IJ/sQUDXPqZRcexFryPDffP1C1SRSaBnx3us=";

          subPackages = [ "cmd" ];
          env.CGO_ENABLED = 0;
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
            "-X main.commit=${commit}"
            "-X main.date=${date}"
          ];

          postInstall = ''
            mv $out/bin/cmd $out/bin/dune
          '';

          meta = {
            description = "Command-line interface for the Dune API";
            homepage = "https://github.com/duneanalytics/cli";
            license = lib.licenses.mit;
            mainProgram = "dune";
          };
        };
      });

      checks = forAllSystems (pkgs: {
        default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      });
    };
}
