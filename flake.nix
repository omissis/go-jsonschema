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
            name = "go-jsonschema-tests";
            src = pkgs.lib.cleanSourceWith {
              src = ./.;
              filter = path: type:
                !(pkgs.lib.hasSuffix "go.work" path)
                && !(pkgs.lib.hasSuffix "go.work.sum" path);
            };

            vendorHash = "sha256-CBxxloy9W9uJq4l2zUrp6VJlu5lNCX55ks8OOWkHDF4=";

            buildPhase = ''
              export HOME=$TMPDIR

              mkdir -p coverage/pkg
              mkdir -p coverage/tests

              echo "Running tests in root module with coverage..."
              go test -v -race -covermode=atomic -coverpkg=./... -cover ./... -args -test.gocoverdir="$PWD/coverage/pkg"

              echo "Running tests in tests module with coverage..."
              go test -v -race -covermode=atomic -coverpkg=./... -cover ./tests -args -test.gocoverdir="$PWD/coverage/tests"

              echo "Generating coverage report..."
              go tool covdata textfmt -i=./coverage/tests,./coverage/pkg -o coverage.out
            '';

            installPhase = ''
              mkdir -p $out
              cp coverage.out $out/ || true
              echo "All tests passed successfully with coverage" > $out/test-results
            '';
          };

          lint-golang = pkgs.buildGoModule {
            name = "lint-golang";
            src = pkgs.lib.cleanSourceWith {
              src = ./.;
              filter = path: type:
                !(pkgs.lib.hasSuffix "go.work" path)
                && !(pkgs.lib.hasSuffix "go.work.sum" path);
            };

            vendorHash = "sha256-CBxxloy9W9uJq4l2zUrp6VJlu5lNCX55ks8OOWkHDF4=";

            nativeBuildInputs = [pkgs.golangci-lint];

            buildPhase = ''
              export HOME=$TMPDIR
              golangci-lint -v run --color=always --config=.rules/.golangci.yml ./...
              golangci-lint -v run --color=always --config=.rules/.golangci.yml tests/*.go
              golangci-lint -v run --color=always --config=.rules/.golangci.yml tests/helpers/*.go
            '';

            installPhase = ''
              mkdir -p $out
              echo "Go linting passed" > $out/result
            '';
          };

          lint-dockerfile = pkgs.stdenv.mkDerivation {
            name = "lint-dockerfile";
            src = ./.;

            nativeBuildInputs = [pkgs.hadolint];

            buildPhase = ''
              find . \
                -type f \
                -name '*Dockerfile*' \
                -not -path './.git/*' \
                -exec hadolint {} \;
            '';

            installPhase = ''
              mkdir -p $out
              echo "Dockerfile linting passed" > $out/result
            '';
          };

          lint-json = pkgs.stdenv.mkDerivation {
            name = "lint-json";
            src = ./.;

            nativeBuildInputs = [pkgs.nodePackages.jsonlint];

            buildPhase = ''
              find . \
                -type f \
                -not -path ".git" \
                -not -path ".github" \
                -not -path ".vscode" \
                -not -path ".idea" \
                -name "*.json" \
                -exec jsonlint -c -q -t '  ' {} \;
            '';

            installPhase = ''
              mkdir -p $out
              echo "JSON linting passed" > $out/result
            '';
          };

          lint-makefile = pkgs.stdenv.mkDerivation {
            name = "lint-makefile";
            src = ./.;

            nativeBuildInputs = [pkgs.checkmake];

            buildPhase = ''
              checkmake --config .rules/checkmake.ini Makefile
            '';

            installPhase = ''
              mkdir -p $out
              echo "Makefile linting passed" > $out/result
            '';
          };

          lint-markdown = pkgs.stdenv.mkDerivation {
            name = "lint-markdown";
            src = ./.;

            nativeBuildInputs = [pkgs.markdownlint-cli2];

            buildPhase = ''
              markdownlint-cli2 "**/*.md" --config ".rules/.markdownlint.yaml"
            '';

            installPhase = ''
              mkdir -p $out
              echo "Markdown linting passed" > $out/result
            '';
          };

          lint-shell = pkgs.stdenv.mkDerivation {
            name = "lint-shell";
            src = ./.;

            nativeBuildInputs = [pkgs.shellcheck];

            buildPhase = ''
              shellcheck -a -o all -s bash -- **/*.sh
            '';

            installPhase = ''
              mkdir -p $out
              echo "Shell linting passed" > $out/result
            '';
          };

          lint-yaml = pkgs.stdenv.mkDerivation {
            name = "lint-yaml";
            src = ./.;

            nativeBuildInputs = [pkgs.yamllint];

            buildPhase = ''
              yamllint -c .rules/yamllint.yaml .
            '';

            installPhase = ''
              mkdir -p $out
              echo "YAML linting passed" > $out/result
            '';
          };

          build-goreleaser = pkgs.stdenv.mkDerivation {
            name = "build-goreleaser";
            src = ./.;

            nativeBuildInputs = [pkgs.go pkgs.goreleaser pkgs.git];

            buildPhase = ''
              export HOME=$TMPDIR
              export GOCACHE=$TMPDIR/go-cache
              export GOPATH=$TMPDIR/go
              export GO_VERSION=$(go version | cut -d ' ' -f 3)

              goreleaser check
              goreleaser release --verbose --snapshot --clean
            '';

            installPhase = ''
              mkdir -p $out
              cp -r dist $out/ || true
              echo "GoReleaser build passed" > $out/result
            '';
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
