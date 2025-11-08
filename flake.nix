{
  description = "go-jsonschema - Generate Go types from JSON Schema";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"];

      perSystem = {
        config,
        pkgs,
        ...
      }: {
        packages = {
          go-jsonschema = pkgs.buildGoModule {
            pname = "go-jsonschema";
            version = "0.0.0-dev";

            src = pkgs.lib.cleanSourceWith {
              src = ./.;
              filter = path: type:
                !(pkgs.lib.hasSuffix "go.work" path)
                && !(pkgs.lib.hasSuffix "go.work.sum" path);
            };

            vendorHash = "sha256-CBxxloy9W9uJq4l2zUrp6VJlu5lNCX55ks8OOWkHDF4=";

            subPackages = ["."];

            ldflags = [
              "-s"
              "-w"
            ];

            meta = with pkgs.lib; {
              description = "Generate Go types from JSON Schema";
              homepage = "https://github.com/atombender/go-jsonschema";
              license = licenses.mit;
              maintainers = [];
            };
          };

          default = config.packages.go-jsonschema;
        };

        checks = {
          go-jsonschema-tests = pkgs.buildGoModule {
            pname = "go-jsonschema-tests";
            version = "0.0.0-dev";

            src = pkgs.lib.cleanSourceWith {
              src = ./.;
              filter = path: type:
                !(pkgs.lib.hasSuffix "go.work" path)
                && !(pkgs.lib.hasSuffix "go.work.sum" path);
            };

            vendorHash = "sha256-CBxxloy9W9uJq4l2zUrp6VJlu5lNCX55ks8OOWkHDF4=";

            doCheck = true;

            checkPhase = ''
              runHook preCheck
              echo "Running tests in root module..."
              go test -v ./...
              runHook postCheck
            '';

            buildPhase = "true";

            installPhase = ''
              runHook preInstall
              mkdir -p $out
              echo "All tests passed successfully" > $out/test-results
              runHook postInstall
            '';

            meta = with pkgs.lib; {
              description = "Test suite for go-jsonschema";
            };
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            go-tools
            golangci-lint
            adr-tools
            goreleaser
            hadolint
            markdownlint-cli2
            shellcheck
            shfmt
            yamllint
            yq-go
            nodePackages.jsonlint
            checkmake
            gofumpt
            jq
          ];

          shellHook = ''
            echo "go-jsonschema development environment"
            echo "Go version: $(go version)"
          '';
        };

        formatter = pkgs.alejandra;
      };
    };
}
