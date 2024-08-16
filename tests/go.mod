module github.com/atombender/go-jsonschema/tests

go 1.22

replace (
	github.com/atombender/go-jsonschema => ../
	github.com/atombender/go-jsonschema/tests/helpers/other => ./helpers/other
)

require (
	github.com/atombender/go-jsonschema v0.15.0
	github.com/atombender/go-jsonschema/tests/helpers/other v0.0.0-20240416182955-738cf5ab9ab4
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/fatih/color v1.16.0 // indirect
	github.com/goccy/go-yaml v1.12.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	golang.org/x/exp v0.0.0-20240808152545-0cdaa3abc0fa // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
)
