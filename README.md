**go-jsonschema is a tool to generate Go data types from [JSON Schema](http://json-schema.org/) definitions.**

This tool generates Go data types and structs that corresponds to definitions in the schema, along with unmarshalling code that validates the input JSON according to the schema's validation rules.

## Badges
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/omissis/go-jsonschema?style=for-the-badge)](https://github.com/omissis/go-jsonschema/releases/latest)
[![GitHub Workflow Status (event)](https://img.shields.io/github/workflow/status/omissis/go-jsonschema/development?style=for-the-badge)](https://github.com/omissis/go-jsonschema/actions?workflow=development)
[![License](https://img.shields.io/github/license/omissis/go-jsonschema?style=for-the-badge)](/LICENSE.md)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/omissis/go-jsonschema?style=for-the-badge)](https://tip.golang.org/doc/go1.20)
[![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/omissis/go-jsonschema?style=for-the-badge)](https://github.com/omissis/go-jsonschema)
[![GitHub repo file count (file type)](https://img.shields.io/github/directory-file-count/omissis/go-jsonschema?style=for-the-badge)](https://github.com/omissis/go-jsonschema)
[![GitHub all releases](https://img.shields.io/github/downloads/omissis/go-jsonschema/total?style=for-the-badge)](https://github.com/omissis/go-jsonschema)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/omissis/go-jsonschema?style=for-the-badge)](https://github.com/omissis/go-jsonschema/commits)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![Codecov branch](https://img.shields.io/codecov/c/github/omissis/go-jsonschema/main.svg?style=for-the-badge)](https://codecov.io/gh/omissis/go-jsonschema)
[![Code Climate maintainability](https://img.shields.io/codeclimate/maintainability/omissis/go-jsonschema?style=for-the-badge)](https://codeclimate.com/github/omissis/go-jsonschema)
[![Go Report Card](https://goreportcard.com/badge/github.com/sighupio/furyctl)](https://goreportcard.com/report/github.com/sighupio/furyctl)

## Installing

* **Download**: Get a release [here](https://github.com/atombender/go-jsonschema/releases).

* **Install from source**: To install with Go 1.16+:

```shell
$ go get github.com/atombender/go-jsonschema/...
$ go install github.com/atombender/go-jsonschema/cmd/gojsonschema@latest
```

* **Install with Bingo**: To install with [Bingo](https://github.com/bwplotka/bingo):

```shell
$ bingo get github.com/atombender/go-jsonschema/cmd/gojsonschema
```

## Usage

At its most basic:

```shell
$ gojsonschema -p main schema.json
```

This will write a Go source file to standard output, declared under the package `main`.

You can generate code for multiple schemas in the same invocation, optionally writing to different files inside different packages:

```shell
$ gojsonschema \
  --schema-package=https://example.com/schema1=github.com/myuser/myproject \
   --schema-output=https://example.com/schema1=schema1.go \
  --schema-package=https://example.com/schema2=github.com/myuser/myproject/stuff \
   --schema-output=https://example.com/schema2=stuff/schema2.go \
  schema1.json schema2.json
```

This will create `schema1.go` (declared as `package myproject`) and `stuff/schema2.go` (declared as `package stuff`). If `schema1.json` refers to `schema2.json` or vice versa, the two Go files will import the other package that it depends on. Note the flag format:

```
--schema-package=https://example.com/schema1=github.com/myuser/myproject \
                 ^                           ^
                 |                           |
                 schema $id                  full import URL
```

## Status

While not finished, go-jsonschema can be used today. Aside from some minor features, only specific validations remain to be fully implemented.

### Validation

- Core ([RFC draft](http://json-schema.org/latest/json-schema-core.html))
  - [x] Data model (§4.2.1)
    - [x] `null`
    - [x] `boolean`
    - [x] `object`
    - [x] `array`
    - [x] `number`
      - [ ] Option to use `json.Number`
    - [x] `string`
  - [ ] Location identifiers (§8.2.3)
    - [x] References against top-level names: `#/$defs/someName`
    - [ ] References against nested names: `#/$defs/someName/$defs/someOtherName`
    - [x] References against top-level names in external files: `myschema.json#/$defs/someName`
    - [ ] References against nested names: `myschema.json#/$defs/someName/$defs/someOtherName`
  - [x] Comments (§9)
- Validation ([RFC draft](http://json-schema.org/latest/json-schema-validation.html))
  - [ ] Schema annotations (§10)
    - [x] `description`
    - [x] `default` (only for struct fields)
    - [ ] `readOnly`
    - [ ] `writeOnly`
    - [ ] ~~`title`~~ (N/A)
    - [ ] ~~`examples`~~ (N/A)
  - [ ] General validation (§6.1)
    - [x] `enum`
    - [x] `type` (single)
    - [x] `type` (multiple; **note**: partial support, limited validation)
    - [ ] `const`
  - [ ] Numeric validation (§6.2)
    - [ ] `multipleOf`
    - [ ] `maximum`
    - [ ] `exclusiveMaximum`
    - [ ] `minimum`
    - [ ] `exclusiveMinimum`
  - [ ] String validation (§6.3)
    - [ ] `maxLength`
    - [ ] `minLength`
    - [ ] `pattern`
  - [ ] Array validation (§6.4)
    - [X] `items`
    - [x] `maxItems`
    - [x] `minItems`
    - [ ] `uniqueItems`
    - [ ] `additionalItems`
    - [ ] `contains`
  - [ ] Object validation (§6.5)
    - [x] `required`
    - [x] `properties`
    - [ ] `patternProperties`
    - [ ] `dependencies`
    - [ ] `propertyNames`
    - [ ] `maxProperties`
    - [ ] `minProperties`
  - [ ] Conditional subschemas (§6.6)
    - [ ] `if`
    - [ ] `then`
    - [ ] `else`
  - [ ] Boolean subschemas (§6.7)
    - [ ] `allOf`
    - [ ] `anyOf`
    - [ ] `oneOf`
    - [ ] `not`
  - [ ] Semantic formats (§7.3)
    - [ ] Dates and times
    - [ ] Email addresses
    - [ ] Hostnames
    - [ ] IP addresses
    - [ ] Resource identifiers
    - [ ] URI-template
    - [ ] JSON pointers
    - [ ] Regex

## License

MIT license. See `LICENSE` file.
