{inputs, ...}: {
  imports = [
    inputs.flake-root.flakeModule
    ./checks.nix
    ./treefmt.nix
    ./nixpkgs.nix
    ./nixos
    ./packages.nix
    ./devshell.nix
    ./dev
  ];
}
