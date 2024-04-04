_: let
  numAgents = 3;
in {
  perSystem = {config, ...}: {
    config.devshells.default = {
      env = [
        {
          name = "VM_DATA_DIR";
          eval = "$PRJ_DATA_DIR/vms";
        }
      ];

      devshell.startup = {
        setup-test-vms.text = ''
          set -euo pipefail

          [ -d $VM_DATA_DIR ] && exit 0
          mkdir -p $VM_DATA_DIR

          for i in {1..${builtins.toString numAgents}}
          do
            OUT="$VM_DATA_DIR/test-vm-$i"
            mkdir -p $OUT
            ssh-keygen -t ed25519 -q -C root@test-vm-$i -N "" -f "$OUT/ssh_host_ed25519_key"
          done
        '';
      };
    };
  };
}
