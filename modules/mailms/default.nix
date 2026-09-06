{
  perSystem =
    { pkgs, self', ... }:
    let
      mailms = self'.packages.mailms;
    in
    {
      checks.mailms = mailms;
      packages.mailms = pkgs.callPackage ./_src { };

      apps.mailms = {
        inherit (mailms) meta;
        program = mailms;
      };
    };
}
