{
  description = "ghx - A TUI dashboard for GitHub";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "ghx";
          version = "0.1.0";
          src = ./.;
          vendorHash = null; # set after first build; use `nix build` to get the hash
          subPackages = [ "cmd" ];

          ldflags = [
            "-s" "-w"
            "-X main.version=0.1.0"
          ];

          meta = with pkgs.lib; {
            description = "A TUI dashboard for GitHub - PRs, Issues, Actions, Notifications";
            homepage = "https://github.com/onnga-wasabi/ghx";
            license = licenses.mit;
            mainProgram = "ghx";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
          ];
        };
      }
    );
}
