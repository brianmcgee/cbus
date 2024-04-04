{
  self,
  inputs,
  ...
}: {
  perSystem = {
    lib,
    pkgs,
    ...
  }: let
    filter = inputs.nix-filter.lib;
  in {
    packages = rec {
      cbus = pkgs.buildGoApplication rec {
        pname = "cbus";
        version = "0.0.1+dev";

        src = filter {
          root = ../.;
          include = [
            "go.mod"
            "go.sum"
            "pkg"
            "internal"
            "cmd"
          ];
        };

        modules = ../gomod2nix.toml;

        ldflags = [
          "-X 'build.Name=${pname}'"
          "-X 'build.Version=${version}'"
        ];

        meta = with lib; {
          description = "CBus - Clustered DBus with NATS";
          homepage = "https://github.com/brianmcgee/cbus";
          license = licenses.mit;
          mainProgram = "cbus";
        };
      };

      default = cbus;
    };
  };

  flake.overlays.default = final: _prev: {
    inherit (self.packages.${final.system}) cbus;
  };
}
