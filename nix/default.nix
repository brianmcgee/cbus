{inputs, ...}: {
  imports = [
    inputs.flake-root.flakeModule
    ./checks.nix
    ./docs.nix
    ./treefmt.nix
    ./nixpkgs.nix
    ./nixos
    ./packages.nix
    ./devshell.nix
    ./dev
  ];
}
