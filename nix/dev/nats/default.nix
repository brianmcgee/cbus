{lib, ...}: {
  perSystem = {pkgs, ...}: let
    config = pkgs.writeTextFile {
      name = "nats.conf";
      text = ''
        ## Default NATS server configuration (see: https://docs.nats.io/running-a-nats-service/configuration)

        ## Host for client connections.
        host: "127.0.0.1"

        ## Port for client connections.
        port: 4222

        ## Port for monitoring
        http_port: 8222

        ## Configuration map for JetStream.
        ## see: https://docs.nats.io/running-a-nats-service/configuration#jetstream
        jetstream {}

        # include paths must be relative so for simplicity we just read in the auth.conf file
        include './auth.conf'
      '';
    };
  in {
    config.process-compose = {
      dev.settings.processes = {
        nats-server = {
          working_dir = "$NATS_HOME";
          command = ''${lib.getExe pkgs.nats-server} -c ./nats.conf -sd ./'';
          readiness_probe = {
            http_get = {
              host = "127.0.0.1";
              port = 8222;
              path = "/healthz";
            };
            initial_delay_seconds = 2;
          };
        };
        nats-setup = {
          depends_on = {
            nats-server.condition = "process_started";
          };
          environment = {
            # ensures contexts are generated in the .data directory
            XDG_CONFIG_HOME = "$PRJ_DATA_DIR";
          };
          command = pkgs.writeShellApplication {
            name = "cbus-setup";
            runtimeInputs = with pkgs; [nsc];
            text = ''
              ACCOUNT="Test"

              for VM_DIR in "$VM_DATA_DIR"/*; do
                  NAME=$(basename "$VM_DIR")
                  NKEY=$(nix run .# -- nkey --public-key-file "$VM_DIR/ssh_host_ed25519_key.pub")

                  # add a user account for the VM
                  nsc add user -a "$ACCOUNT" \
                      -k "$NKEY" -n "$NAME" \
                      --allow-sub "dbus.agent.$NKEY.>" \
                      --allow-sub "dbus.broadcast.>" \
                      --allow-pub "dbus.signals.$NKEY.>" \
                      --allow-sub "\$SRV.>" \
                      --allow-pub "_INBOX.>"

                  # capture the JWT
                  nsc describe user -a "$ACCOUNT" -n "$NAME" -R > "$VM_DIR/user.jwt"
              done

              # push any account/user changes to the server
              nsc push
            '';
          };
        };
      };
    };

    config.devshells.default = {
      devshell.startup = {
        setup-nats = {
          deps = ["setup-test-vms"];
          text = ''
            set -euo pipefail

            # we only setup the data dir if it doesn't exist
            # to refresh simply delete the directory and run `direnv reload`
            [ -d $NSC_HOME ] && exit 0

            mkdir -p $NSC_HOME

            # initialise nsc state

            nsc init -n CBus --dir $NSC_HOME
            nsc edit operator \
                --service-url nats://localhost:4222 \
                --account-jwt-server-url nats://localhost:4222

            # setup server config

            mkdir -p $NATS_HOME
            cp ${config} "$NATS_HOME/nats.conf"
            nsc generate config --nats-resolver --config-file "$NATS_HOME/auth.conf"

            # Add a Test account
            nsc add account -n Test --deny-pubsub '>'
            nsc edit account -n Test \
                --js-mem-storage -1 \
                --js-disk-storage -1 \
                --js-streams -1 \
                --js-consumer -1

            # Add a Test Admin user
            nsc add user -a Test -n Admin --allow-pubsub '>'
            nsc generate context -a Test -u Admin --context TestAdmin
          '';
        };
      };

      env = [
        {
          name = "NATS_HOME";
          eval = "$PRJ_DATA_DIR/nats";
        }
        {
          name = "NSC_HOME";
          eval = "$PRJ_DATA_DIR/nsc";
        }
        {
          name = "NKEYS_PATH";
          eval = "$NSC_HOME";
        }
        {
          name = "NATS_JWT_DIR";
          eval = "$PRJ_DATA_DIR/nats/jwt";
        }
      ];

      packages = [
        pkgs.nkeys
        pkgs.nats-top
      ];

      commands = let
        category = "nats";
      in [
        {
          inherit category;
          name = "nsc";
          command = ''XDG_CONFIG_HOME=$PRJ_DATA_DIR ${pkgs.nsc}/bin/nsc -H "$NSC_HOME" "$@"'';
        }
        {
          inherit category;
          name = "nats";
          command = ''XDG_CONFIG_HOME=$PRJ_DATA_DIR ${pkgs.natscli}/bin/nats "$@"'';
        }
      ];
    };
  };
}
