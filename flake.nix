{
  description = "Nix flake for memri project";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-26.05";
  };

  outputs =
    { self, nixpkgs }:
    let
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
      myShell = pkgs.mkShell {
        buildInputs = [
          pkgs.go
          pkgs.go-task
          pkgs.goose
          pkgs.air
          pkgs.bun
          pkgs.nodejs_22
          pkgs.dotenvx
        ];
        GOFLAGS = "-mod=readonly";
      };
    in
    {
      devShells = {
        default = myShell;
        "x86_64-linux" = {
          default = myShell;
        };
      };
    };
}
