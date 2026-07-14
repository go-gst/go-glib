{
  description = "Golang+gstreamer dev environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
    ...
  }: let
    systems = ["x86_64-linux" "aarch64-darwin"];
    forAllSystems = nixpkgs.lib.genAttrs systems;
  in {
    formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.alejandra);
    devShells = forAllSystems (system: let
      pkgs = import nixpkgs {
        inherit system;
      };
    in {
      default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          go_latest
          go-tools # staticcheck & co.
          pkg-config

          glib
        ];

        GO111MODULE = "on";
        CGO_ENABLED = "1";

        # needed for running delve
        # https://github.com/go-delve/delve/issues/3085
        # https://nixos.wiki/wiki/C#Hardening_flags
        # hardeningDisable = ["all"];

        # print the go version and gstreamer version on shell startup
        shellHook = ''
          ${pkgs.go_latest}/bin/go version
        '';
      };
    });
  };
}
