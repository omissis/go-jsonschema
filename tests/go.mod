module github.com/atombender/go-jsonschema/tests

go 1.24.0

replace (
	github.com/atombender/go-jsonschema => ../
	github.com/atombender/go-jsonschema/tests/helpers/other => ./helpers/other
)

require (
	github.com/atombender/go-jsonschema v0.22.0
	github.com/atombender/go-jsonschema/tests/helpers/other v0.0.0-20260228134922-cb7689305858
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/google/go-cmp v0.7.0
	github.com/stretchr/testify v1.11.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sanity-io/litter v1.5.8 // indirect
	github.com/sosodev/duration v1.4.0 // indirect
)
