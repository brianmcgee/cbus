{
  lib,
  config,
  inputs,
  pkgs,
  ...
}: let
  cfg = config.services.cbus-agent;
in {
  options.services.cbus-agent = with lib; {
    enable = mkEnableOption (mdDoc "Enable CBus agent");
    package = mkOption {
      type = types.package;
      default = inputs.cbus.packages.${pkgs.system}.cbus;
      description = mdDoc "Package to use for cbus agent.";
    };
    nats = {
      url = mkOption {
        type = types.str;
        example = "nats://localhost:4222";
        description = mdDoc "NATS server url.";
      };
      jwtFile = mkOption {
        type = types.path;
        example = "/mnt/shared/user.jwt";
        description = mdDoc "Path to a file containing a NATS JWT token.";
      };
      hostKeyFile = mkOption {
        type = types.path;
        default = "/etc/ssh/ssh_host_ed25519_key";
        example = "/etc/ssh/ssh_host_ed25519_key";
        description = mdDoc "Path to an ed25519 host key file";
      };
    };
    logLevel = mkOption {
      type = types.int;
      default = 1;
      example = "1";
      description = mdDoc "Selects the logging verbosity. 0 = warn, 1 = info, >=2 = debug.";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.cbus-agent = {
      after = ["network.target"];
      wantedBy = ["sysinit.target"];

      description = "CBus Agent";
      startLimitIntervalSec = 0;

      environment = lib.filterAttrs (_: v: v != null) {
        NATS_URL = cfg.nats.url;
        NATS_HOST_KEY_FILE = cfg.nats.hostKeyFile;
        NATS_JWT_FILE = cfg.nats.jwtFile;
        LOG_LEVEL = builtins.toString cfg.logLevel;
      };

      serviceConfig = with lib; {
        Restart = mkDefault "on-failure";
        RestartSec = 1;

        User = "root";
        StateDirectory = "cbus-agent";
        ExecStart = "${cfg.package}/bin/cbus agent";
      };
    };
  };
}
