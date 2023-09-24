{
  description = "A flake for building gnut";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: 
  flake-utils.lib.eachDefaultSystem (system:
    let
        name = "gnut";
        pkgs = import nixpkgs {
            inherit system;
        };
        version = "0.1"; 
        buildGoModule = pkgs.buildGo121Module;
        bin = buildGoModule ({
            inherit version;
            src = ./.;
            pname = name;
            vendorHash = "sha256-+zKY4KNUipSmu6108n7LThRf/wdZJuoARFkMkSzCyNw=";
            CGO_ENABLED = "0";

            meta = with pkgs.lib; {
                description = "A interface for managing your NUT server";
                homepage = "github.com/ethanholz/gnut";
                license = licenses.mit;
                maintainers = [ "ethanholz" ];
            };
        });
        dockerImage = pkgs.dockerTools.buildImage {
            name = "${name}";
            tag = "latest";
            created = "now";
            copyToRoot = [ pkgs.cacert ];
            config = {
                Cmd = [ "${bin}/bin/gnut" ];
            };  
        };
    in
    {
        devShells = {
            default = pkgs.mkShell {
                nativeBuildInputs = (with pkgs; [go_1_21]);
                buildInputs = (with pkgs; [gopls dive]);
            };
        };
        packages = rec {
            default = gnut;
            gnut = bin;
            docker = dockerImage;
        };
  });
    # packages.x86_64-linux.hello = nixpkgs.legacyPackages.x86_64-linux.hello;
    #
    # packages.x86_64-linux.default = self.packages.x86_64-linux.hello;

}
