{ stdenv, fetchurl, autoPatchelfHook }:

stdenv.mkDerivation rec {
  pname = "of-watchdog";
  version = "0.8.0";

  src = fetchurl {
    url =
      "https://github.com/openfaas/of-watchdog/releases/download/${version}/of-watchdog";
    sha256 = "0pkqviv82gpqv15a4dmawzdwb1041lnlpaqxxpxy1z2mq7y673f7";
  };

  dontUnpack = true;
  dontBuild = true;

  nativeBuildInputs = [ autoPatchelfHook ];

  installPhase = ''
    mkdir -p $out/bin
    install -m755 $src $out/bin/of-watchdog
  '';
}
