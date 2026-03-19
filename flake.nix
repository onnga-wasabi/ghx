{
  description = "ghx - A keyboard-driven TUI for GitHub";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      version = "0.2.0";
      mkGhx = pkgs: pkgs.buildGoModule {
        pname = "ghx";
        inherit version;
        src = ./.;
        vendorHash = "sha256-VBw8nOlFkyKuGB+3ZFejQZxQ7PYgYvRJpFw4iFZXBv4=";
        subPackages = [ "cmd" ];

        postInstall = ''
          mv $out/bin/cmd $out/bin/ghx 2>/dev/null || true
        '';

        ldflags = [
          "-s" "-w"
          "-X main.version=${version}"
        ];

        meta = with pkgs.lib; {
          description = "A keyboard-driven TUI for GitHub — PRs, Issues, Actions, Notifications";
          homepage = "https://github.com/onnga-wasabi/ghx";
          license = licenses.mit;
          mainProgram = "ghx";
        };
      };
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = mkGhx pkgs;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
          ];
        };
      }
    ) // {
      overlays.default = final: prev: {
        ghx = mkGhx final;
      };
    };
}
