module github.com/atombender/go-jsonschema/tests

go 1.21.2

replace (
	github.com/atombender/go-jsonschema => ../
	github.com/atombender/go-jsonschema/tests/helpers/other => ./helpers/other
)

require (
	github.com/atombender/go-jsonschema v0.0.0-00010101000000-000000000000
	github.com/atombender/go-jsonschema/tests/helpers/other v0.0.0-00010101000000-000000000000
	github.com/magiconair/properties v1.8.7
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/fatih/color v1.13.0 // indirect
	github.com/goccy/go-yaml v1.11.2 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	golang.org/x/exp v0.0.0-20231005195138-3e424a577f31 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)
