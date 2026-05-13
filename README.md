**go-jsonschema is a tool to generate Go data types from [JSON Schema](http://json-schema.org/) definitions.**

This tool generates Go data types and structs that corresponds to definitions in the schema,
along with unmarshalling code that validates the input JSON according to the schema's validation rules.

## Badges

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tuotuoxp/go-jsonschema?style=flat)](https://github.com/tuotuoxp/go-jsonschema/releases/latest)
[![GitHub Workflow Status (event)](https://img.shields.io/github/actions/workflow/status/tuotuoxp/go-jsonschema/development.yaml?style=flat)](https://github.com/tuotuoxp/go-jsonschema/actions?workflow=development)
[![License](https://img.shields.io/github/license/tuotuoxp/go-jsonschema?style=flat)](/LICENSE.md)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tuotuoxp/go-jsonschema?style=flat)](https://tip.golang.org/doc/go1.25)
[![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/tuotuoxp/go-jsonschema?style=flat)](https://github.com/tuotuoxp/go-jsonschema)
[![GitHub repo file count (file type)](https://img.shields.io/github/directory-file-count/tuotuoxp/go-jsonschema?style=flat)](https://github.com/tuotuoxp/go-jsonschema)
[![GitHub all releases](https://img.shields.io/github/downloads/tuotuoxp/go-jsonschema/total?style=flat)](https://github.com/tuotuoxp/go-jsonschema)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/tuotuoxp/go-jsonschema?style=flat)](https://github.com/tuotuoxp/go-jsonschema/commits)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=flat)](https://conventionalcommits.org)
[![Codecov](https://img.shields.io/codecov/c/gh/tuotuoxp/go-jsonschema?style=flat&token=lPWlXd3MVK)](https://codecov.io/gh/tuotuoxp/go-jsonschema)
[![Code Climate maintainability](https://img.shields.io/codeclimate/maintainability/tuotuoxp/go-jsonschema?style=flat)](https://codeclimate.com/github/tuotuoxp/go-jsonschema)
[![Go Report Card](https://goreportcard.com/badge/github.com/tuotuoxp/go-jsonschema)](https://goreportcard.com/report/github.com/tuotuoxp/go-jsonschema)

## Installing

* **Download**: Get a release [here](https://github.com/tuotuoxp/go-jsonschema/releases).

* **Install from source**: To install with Go 1.25+:

```shell
go get github.com/tuotuoxp/go-jsonschema/...
go install github.com/tuotuoxp/go-jsonschema@latest
```

* **Install with Brew**: To install with [Homebrew](https://brew.sh):

```shell
brew tap tuotuoxp/go-jsonschema
brew install go-jsonschema
```

## Contributing

This project makes use of [go workspaces](https://go.dev/ref/mod#workspaces) in order to ease testing of the
generated code during development while keeping the codebase as tidy and maintainable as possible.
It's an unusual choice, but it allows to not only test the code-generation logic, but also the generated code itself.

## Usage

At its most basic:

```shell
go-jsonschema -p main schema.json
```

This will write a Go source file to standard output, declared under the package `main`.

You can generate code for multiple schemas in one run, optionally writing to different files inside different packages:

```shell
$ go-jsonschema \
  --schema-package=https://example.com/schema1=github.com/myuser/myproject \
   --schema-output=https://example.com/schema1=schema1.go \
  --schema-package=https://example.com/schema2=github.com/myuser/myproject/stuff \
   --schema-output=https://example.com/schema2=stuff/schema2.go \
  schema1.json schema2.json
```

This will create `schema1.go` (declared as `package myproject`) and `stuff/schema2.go` (declared as `package stuff`).
If `schema1.json` refers to `schema2.json` or viceversa, the two Go files will import the other depended-on package.
Note the flag format:

```text
--schema-package=https://example.com/schema1=github.com/myuser/myproject \
                 ^                           ^
                 |                           |
                  schema $id                  full import URL
```

### Schema extension: `x-go-oneof-envelope`

`x-go-oneof-envelope` is a first-stage extension for discriminator-driven envelope
shapes such as `{ "type": "...", "value": { ... } }`.

- Use it on the envelope payload field (for example, `value`) that contains `oneOf`.
- `discriminator` is the sibling field name (for example, `type`) used for routing.
- `mapping` keys are discriminator values, and mapping values must match the resolved
  branch `title` values.
- `oneOf` branches may be `$ref` or inline, but each branch must resolve to a unique
  `title`.
- Current supported usage assumes one such envelope field per object.

Example:

```json
{
  "type": "object",
  "required": ["type", "value"],
  "properties": {
    "type": {
      "type": "string",
      "enum": ["a", "b"]
    },
    "value": {
      "title": "Payload",
      "oneOf": [
        { "$ref": "#/$defs/AValue" },
        {
          "title": "BValue",
          "type": "object",
          "required": ["sub_b"],
          "properties": {
            "sub_b": { "type": "integer" }
          }
        }
      ],
      "x-go-oneof-envelope": {
        "discriminator": "type",
        "mapping": {
          "a": "AValue",
          "b": "BValue"
        }
      }
    }
  },
  "$defs": {
    "AValue": {
      "title": "AValue",
      "type": "object",
      "required": ["sub_a"],
      "properties": {
        "sub_a": { "type": "string" }
      }
    }
  }
}
```

### Regenerating tests' golden files

It sometimes happen that new features or bug fixes to the library require regenerating the tests' golden files, here's how to do it:

```
export OVERWRITE_EXPECTED_GO_FILE="true"
make test
```

### Special types

In a few cases, special types are used to help with serializing/deserializing
data frrom JSON. Namely a custom types is provided for the following semantic
types:

* `SerializableDate`
* `SerializableTime`

These types are needed because there is no native type provided by Go which
properly handles them.

## Status

While not finished, go-jsonschema can be used today. Aside from some minor features,
only specific validations remain to be fully implemented.

### Validation

* Core ([RFC draft](http://json-schema.org/latest/json-schema-core.html))
  * [x] Data model (§4.2.1)
    * [x] `null`
    * [x] `boolean`
    * [x] `object`
    * [x] `array`
    * [x] `number`
      * [ ] Option to use `json.Number`
    * [x] `string`
  * [ ] Location identifiers (§8.2.3)
    * [x] References against top-level names: `#/$defs/someName`
    * [ ] References against nested names: `#/$defs/someName/$defs/someOtherName`
    * [x] References against top-level names in external files: `myschema.json#/$defs/someName`
    * [ ] References against nested names: `myschema.json#/$defs/someName/$defs/someOtherName`
  * [x] Comments (§9)
* Validation ([RFC draft](http://json-schema.org/latest/json-schema-validation.html))
  * [ ] Schema annotations (§10)
    * [x] `description`
    * [x] `default` (only for struct fields)
    * [ ] `readOnly`
    * [ ] `writeOnly`
    * [ ] ~~`title`~~ (N/A)
    * [ ] ~~`examples`~~ (N/A)
  * [ ] General validation (§6.1)
    * [x] `enum`
    * [x] `type` (single)
    * [x] `type` (multiple; **note**: partial support, limited validation)
    * [ ] `const`
  * [X] Numeric validation (§6.2)
    * [X] `multipleOf`
    * [X] `maximum`
    * [X] `exclusiveMaximum`
    * [X] `minimum`
    * [X] `exclusiveMinimum`
  * [X] String validation (§6.3)
    * [X] `maxLength`
    * [X] `minLength`
    * [X] `pattern`
  * [ ] Array validation (§6.4)
    * [X] `items`
    * [x] `maxItems`
    * [x] `minItems`
    * [ ] `uniqueItems`
    * [ ] `additionalItems`
    * [ ] `contains`
  * [ ] Object validation (§6.5)
    * [x] `required`
    * [x] `properties`
    * [ ] `patternProperties`
    * [ ] `dependencies`
    * [ ] `propertyNames`
    * [ ] `maxProperties`
    * [ ] `minProperties`
  * [ ] Conditional subschemas (§6.6)
    * [ ] `if`
    * [ ] `then`
    * [ ] `else`
  * [ ] Boolean subschemas (§6.7)
    * [ ] `allOf`
    * [ ] `anyOf`
    * [ ] `oneOf`
    * [ ] `not`
  * [ ] Semantic formats (§7.3)
    * [x] Dates and times
    * [ ] Email addresses
    * [ ] Hostnames
    * [ ] IP addresses
    * [ ] Resource identifiers
    * [ ] URI-template
    * [ ] JSON pointers
    * [ ] Regex

## License

MIT license. See `LICENSE` file.
