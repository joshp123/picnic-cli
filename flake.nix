{
  description = "Picnic grocery shopping CLI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      pluginFor = system: {
        name = "picnic";
        skills = [ ./skills/picnic ];
        packages = [ self.packages.${system}.default ];
        needs = {
          requiredEnv = [ "PICNIC_AUTH_FILE" ];
        };
      };
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        packages.default = pkgs.buildGoModule {
          pname = "picnic-cli";
          version = "0.1.0";
          src = ./.;
          go = pkgs.go_1_25;
          vendorHash = null;
          postInstall = ''
            ln -s $out/bin/picnic-cli $out/bin/picnic
          '';
        };

        devShells.default = pkgs.mkShell {
          buildInputs = [ pkgs.go pkgs.gopls ];
        };
      }) // {
        clawdbotPlugin = pluginFor builtins.currentSystem;
      };
}
