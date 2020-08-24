{ pkgs }:

let
  of-watchdog = pkgs.callPackage ./nix/of-watchdog.nix { };
  youtube_adguard = pkgs.callPackage ./default.nix { };
in pkgs.dockerTools.buildImage {
  name = "youtube_adguard";
  tag = "v0.1.0";

  config = {
    WorkingDir = "/";
    Env = [
      "fprocess=${youtube_adguard}/bin/youtube_adguard"
      "mode=http"
      "upstream_url=http://127.0.0.1:8082"
    ];
    Cmd = "${of-watchdog}/bin/of-watchdog";

    contents = [ of-watchdog youtube_adguard pkgs.cacert ];
  };
}
