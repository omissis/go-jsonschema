{
  description = "go-jsonschema - Generate Go types from JSON Schema";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    treefmt-nix.inputs.nixpkgs.follows = "nixpkgs";
    git-hooks.url = "github:cachix/git-hooks.nix";
    git-hooks.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        inputs.treefmt-nix.flakeModule
        inputs.git-hooks.flakeModule
      ];

      systems = ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"];

      perSystem = {
        config,
        pkgs,
        ...
      }: let
        cleanSrc = pkgs.lib.cleanSourceWith {
          src = ./.;
          filter = path: type:
            !(pkgs.lib.hasSuffix "go.work" path)
            && !(pkgs.lib.hasSuffix "go.work.sum" path);
        };

        makePackage = goPackage: let
          buildGoModule = pkgs.buildGoModule.override {go = goPackage;};
        in
          buildGoModule {
            pname = "go-jsonschema";
            version = "0.0.0-dev";
            src = cleanSrc;
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

        makeTests = goPackage: let
          buildGoModule = pkgs.buildGoModule.override {go = goPackage;};
        in
          buildGoModule {
            name = "go-jsonschema-tests";
            src = cleanSrc;
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
      in {
        packages = {
          go-jsonschema-go124 = makePackage pkgs.go_1_24;
          go-jsonschema-go125 = makePackage pkgs.go;
          default = makePackage pkgs.go;

          test-ci = pkgs.writeShellApplication {
            name = "test-ci";
            runtimeInputs = [pkgs.act];
            text = ''
              exec act -W .github/workflows/nix.yaml -P ubuntu-24.04=catthehacker/ubuntu:act-24.04 "$@"
            '';
          };
        };

        checks = {
          tests-go124 = makeTests pkgs.go_1_24;
          tests-go125 = makeTests pkgs.go;

          lint-golang = let
            buildGoModule = pkgs.buildGoModule.override {go = pkgs.go;};
          in
            buildGoModule {
              name = "lint-golang";
              src = cleanSrc;
              vendorHash = "sha256-CBxxloy9W9uJq4l2zUrp6VJlu5lNCX55ks8OOWkHDF4=";
              nativeBuildInputs = [pkgs.golangci-lint];
              buildPhase = ''
                export HOME=$TMPDIR
                golangci-lint -v run --modules-download-mode vendor --color=always --config=.rules/.golangci.yml ./... --skip-dirs tests
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

          build-goreleaser = let
            buildGoModule = pkgs.buildGoModule.override {go = pkgs.go;};
          in
            buildGoModule {
              name = "build-goreleaser";
              src = cleanSrc;
              vendorHash = "sha256-CBxxloy9W9uJq4l2zUrp6VJlu5lNCX55ks8OOWkHDF4=";
              nativeBuildInputs = [pkgs.goreleaser pkgs.git];

              buildPhase = ''
                export HOME=$TMPDIR
                export GO_VERSION=$(go version | cut -d ' ' -f 3)

                goreleaser check
                goreleaser release --skip=before,docker --verbose --snapshot --clean
              '';

              installPhase = ''
                mkdir -p $out
                cp -r dist $out/ || true
                echo "GoReleaser build passed" > $out/result
              '';
            };
        };

        treefmt = {
          projectRootFile = "flake.nix";

          programs = {
            alejandra.enable = true;
            # Disabled to avoid modifying .go files in tests/data
            # gofmt.enable = true;
            # gofumpt.enable = true;
          };

          settings.global.excludes = [
            ".direnv/*"
            "result"
            ".git/*"
          ];

          settings.formatter = {
            shfmt = {
              command = "${pkgs.shfmt}/bin/shfmt";
              options = ["-i" "2" "-ci" "-sr" "-w"];
              includes = ["*.sh"];
            };

            # Temporarily disabled - yq keeps rewriting files even without changes
            # yq = {
            #   command = pkgs.writeShellApplication {
            #     name = "format-yaml";
            #     runtimeInputs = [pkgs.yq-go];
            #     text = ''
            #       for file in "$@"; do
            #         yq eval -P -I 2 -i "$file"
            #       done
            #     '';
            #   };
            #   includes = ["*.yaml" "*.yml"];
            # };

            # Disabled to avoid modifying .json files in tests/data
            # jq = {
            #   command = pkgs.writeShellApplication {
            #     name = "format-json";
            #     text = ''
            #       for file in "$@"; do
            #         ${pkgs.jq}/bin/jq -M . "$file" > "$file.tmp" && mv "$file.tmp" "$file"
            #       done
            #     '';
            #   };
            #   includes = ["*.json"];
            # };

            # Disabled to avoid modifying .go files in tests/data
            # goimports = {
            #   command = "${pkgs.gotools}/bin/goimports";
            #   options = ["-w" "-local" "github.com/atombender"];
            #   includes = ["*.go"];
            #   excludes = ["vendor/*" "tests/data/*"];
            # };

            markdownlint-cli2 = {
              command = "${pkgs.markdownlint-cli2}/bin/markdownlint-cli2";
              options = ["--config" ".rules/.markdownlint.yaml" "--fix"];
              includes = ["*.md"];
              excludes = [".direnv/*" "result/*"];
            };
          };
        };

        pre-commit = {
          check.enable = true;
          settings.hooks = {
            treefmt = {
              enable = true;
              package = config.treefmt.build.wrapper;
              # Don't use --fail-on-change, just format the files
              entry = "${config.treefmt.build.wrapper}/bin/treefmt";
              pass_filenames = false;
            };
          };
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs;
            [
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
              config.treefmt.build.wrapper
              act
              config.pre-commit.settings.package
            ]
            ++ config.pre-commit.settings.enabledPackages;

          shellHook = ''
            ${config.pre-commit.installationScript}
            echo "go-jsonschema development environment"
            echo "Go version: $(go version)"
          '';
        };

        formatter = config.treefmt.build.wrapper;
      };
    };
}
