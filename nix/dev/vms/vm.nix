{
  config,
  inputs,
  modulesPath,
  ...
}: {
  imports = [
    inputs.cbus.nixosModules.agent
    "${toString modulesPath}/virtualisation/qemu-vm.nix"
  ];

  services.getty.autologinUser = "root";

  services = {
    cbus-agent = {
      enable = true;
      nats = {
        url = "nats://10.0.2.2";
        jwtFile = "/mnt/shared/user.jwt";
      };
    };

    openssh = {
      enable = true;
      settings = {
        PermitRootLogin = "yes";
      };
    };
  };

  virtualisation = let
    inherit (config.networking) hostName;
  in {
    graphics = false;
    diskImage = "$VM_DATA_DIR/${hostName}/disk.qcow2";
    writableStoreUseTmpfs = false;
    sharedDirectories = {
      config = {
        source = "$VM_DATA_DIR/${hostName}";
        target = "/mnt/shared";
      };
    };
  };

  system.stateVersion = config.system.nixos.version;

  system.activationScripts = {
    # replace host key with pre-generated one
    host-key.text = ''
      rm -f /etc/ssh/ssh_host_ed25519_key*
      cp /mnt/shared/ssh_host_ed25519_key /etc/ssh/ssh_host_ed25519_key
      cp /mnt/shared/ssh_host_ed25519_key.pub /etc/ssh/ssh_host_ed25519_key.pub

      chmod 600 /etc/ssh/ssh_host_ed25519_key
      chmod 644 /etc/ssh/ssh_host_ed25519_key.pub
    '';
  };
}
