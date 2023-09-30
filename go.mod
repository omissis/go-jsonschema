module github.com/atombender/go-jsonschema

go 1.21

require (
	github.com/atombender/go-jsonschema/tests/data v0.0.0-00010101000000-000000000000
	github.com/goccy/go-yaml v1.11.2
	github.com/magiconair/properties v1.8.7
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/sanity-io/litter v1.5.5
	github.com/spf13/cobra v1.7.0
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/fatih/color v1.13.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/atombender/go-jsonschema/tests/data => ./tests/data
