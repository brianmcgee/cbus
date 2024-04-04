{inputs, ...}: {
  perSystem = {system, ...}: {
    # customize nixpkgs instance
    _module.args.pkgs = import inputs.nixpkgs {
      inherit system;
      overlays = [
        # adds buildGoApplication
        inputs.gomod2nix.overlays.default
      ];
    };
  };
}
