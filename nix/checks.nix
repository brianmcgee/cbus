{
  lib,
  self,
  inputs,
  ...
}: {
  flake.githubActions = let
    platforms = {
      "x86_64-linux" = ["ubuntu-latest"];
    };
  in
    inputs.github-actions.lib.mkGithubMatrix {
      inherit platforms;
      checks = lib.getAttrs ["x86_64-linux"] self.checks;
    };

  perSystem = {self', ...}: {
    checks = with lib; mapAttrs' (n: nameValuePair "package-${n}") self'.packages;
  };
}
