{ lib, pkgs }:
pkgs.pkgsStatic.buildGoModule {
  doCheck = true;
  name = "mailms";
  src = ./mailms;
  vendorHash = "sha256-ivlvT5+XUYflEZxcieQr1E/t1R+75Oblb/Ne5wmJenU=";

  checkPhase = ''
    runHook preCheck
    go vet ./...
    runHook postCheck
  '';

  meta = {
    description = "mailms - Mail Microservice";
    homepage = "https://github.com/andrieee44/mailms";
    license = lib.licenses.agpl3Plus;
    mainProgram = "mailms";
  };
}
