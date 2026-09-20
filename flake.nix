{
  description = "Nix flake for memri project";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
  };

  outputs = { self, nixpkgs }: let
    pkgs = nixpkgs.legacyPackages.x86_64-linux;
    myShell = pkgs.mkShell {
      buildInputs = [
        pkgs.go
        pkgs.go-task
        pkgs.bun
        pkgs.nodejs_22
      ];
      GOFLAGS = "-mod=readonly";
      path = "${pkgs.coreutils}/bin";
    };
  in {
    devShells = {
      default = myShell;
      "x86_64-linux" = {
        default = myShell;
      };
    };
  };
}
