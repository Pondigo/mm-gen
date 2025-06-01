{
  description = "MM Go Agent - Diagram generator CLI tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "mm-go-agent";
          version = "0.1.0";
          
          src = ./.;
          
          # You'll need to calculate this - see instructions below
          vendorHash = "sha256-NKdtPVQx4QLYSkNs1mHruthRD7KHM+iTZ2APilm8SsU="; # Correct vendor hash from Nix
          
          # Since main.go is in cmd/, we need to specify the subdirectory
          subPackages = [ "cmd" ];
          
          # Rename the binary from "cmd" to "mm-agent"
          postInstall = ''
            mv $out/bin/cmd $out/bin/mm-agent
          '';
          
          meta = with pkgs.lib; {
            description = "Diagram generator CLI tool using mermaid-go";
            homepage = "https://github.com/yourusername/mm-go-agent";
            license = licenses.mit;
            maintainers = with maintainers; [ ];
            mainProgram = "mm-agent";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
          ];
          
          shellHook = ''
            echo "Welcome to mm-go-agent development environment"
            echo "Go version: $(go version)"
          '';
        };
      });
} 