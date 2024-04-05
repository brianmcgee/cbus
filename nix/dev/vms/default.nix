{
  self,
  inputs,
  lib,
  ...
}: {
  perSystem = {
    config,
    pkgs,
    system,
    ...
  }: let
    numAgents = 3;
  in {
    config.nixosConfigurations =
      builtins.listToAttrs
      (builtins.map
        (id: let
          name = "test-machine-${builtins.toString id}";
        in
          lib.nameValuePair name (lib.nixosSystem {
            inherit pkgs;
            specialArgs = {
              # pass in self as cbus input
              inputs = inputs // {cbus = self;};
            };
            modules = [
              ./vm.nix
              {networking.hostName = name;}
            ];
          }))
        (lib.range 1 numAgents));

    config.devshells.default = {
      env = [
        {
          name = "VM_DATA_DIR";
          eval = "$PRJ_DATA_DIR/vms";
        }
      ];

      devshell.startup = {
        setup-test-machines.text = ''
          set -euo pipefail

          [ -d $VM_DATA_DIR ] && exit 0
          mkdir -p $VM_DATA_DIR

          for i in {1..${builtins.toString numAgents}}
          do
            OUT="$VM_DATA_DIR/test-machine-$i"
            mkdir -p $OUT
            ssh-keygen -t ed25519 -q -C root@test-machine-$i -N "" -f "$OUT/ssh_host_ed25519_key"
          done
        '';
      };

      commands = [
        {
          category = "development";
          help = "run a test vm";
          name = "run-test-machine";
          command = "nix run .#nixosConfigurations.${system}_test-machine-$1.config.system.build.vm";
        }
      ];
    };

    config.process-compose = {
      dev.settings.processes = let
        mkAgentProcess = id: {
          command = "run-test-machine ${builtins.toString id}";
        };
        configs =
          map
          (id: lib.nameValuePair "test-machine-${builtins.toString id}" (mkAgentProcess id))
          (lib.range 1 numAgents);
      in
        builtins.listToAttrs configs;
    };
  };
}
