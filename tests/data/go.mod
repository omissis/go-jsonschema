module github.com/atombender/go-jsonschema/tests/data

go 1.21

replace (
	github.com/atombender/go-jsonschema/tests/helpers/other => ../helpers/other
	github.com/atombender/go-jsonschema/tests/helpers/schema => ../helpers/schema
)

require (
	github.com/atombender/go-jsonschema/tests/helpers/other v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)
