module github.com/atombender/go-jsonschema/tests

go 1.21

replace (
	github.com/atombender/go-jsonschema => ../
	github.com/atombender/go-jsonschema/tests/helpers/other => ./helpers/other
)

require (
	github.com/atombender/go-jsonschema v0.0.0-00010101000000-000000000000
	github.com/atombender/go-jsonschema/tests/helpers/other v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/fatih/color v1.16.0 // indirect
	github.com/goccy/go-yaml v1.11.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	golang.org/x/exp v0.0.0-20240112132812-db7319d0e0e3 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
)
