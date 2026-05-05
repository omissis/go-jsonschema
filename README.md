**go-jsonschema is a tool to generate Go data types from [JSON Schema](http://json-schema.org/) definitions.**

This tool generates Go data types and structs that corresponds to definitions in the schema,
along with unmarshalling code that validates the input JSON according to the schema's validation rules.

## Badges

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/omissis/go-jsonschema?style=flat)](https://github.com/omissis/go-jsonschema/releases/latest)
[![GitHub Workflow Status (event)](https://img.shields.io/github/actions/workflow/status/omissis/go-jsonschema/development.yaml?style=flat)](https://github.com/omissis/go-jsonschema/actions?workflow=development)
[![License](https://img.shields.io/github/license/omissis/go-jsonschema?style=flat)](/LICENSE.md)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/omissis/go-jsonschema?style=flat)](https://tip.golang.org/doc/go1.25)
[![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/omissis/go-jsonschema?style=flat)](https://github.com/omissis/go-jsonschema)
[![GitHub repo file count (file type)](https://img.shields.io/github/directory-file-count/omissis/go-jsonschema?style=flat)](https://github.com/omissis/go-jsonschema)
[![GitHub all releases](https://img.shields.io/github/downloads/omissis/go-jsonschema/total?style=flat)](https://github.com/omissis/go-jsonschema)
[![GitHub commit activity](https://img.shields.io/github/commit-activity/y/omissis/go-jsonschema?style=flat)](https://github.com/omissis/go-jsonschema/commits)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=flat)](https://conventionalcommits.org)
[![Codecov](https://img.shields.io/codecov/c/gh/omissis/go-jsonschema?style=flat&token=lPWlXd3MVK)](https://codecov.io/gh/omissis/go-jsonschema)
[![Code Climate maintainability](https://img.shields.io/codeclimate/maintainability/omissis/go-jsonschema?style=flat)](https://codeclimate.com/github/omissis/go-jsonschema)
[![Go Report Card](https://goreportcard.com/badge/github.com/omissis/go-jsonschema)](https://goreportcard.com/report/github.com/omissis/go-jsonschema)

## Installing

* **Download**: Get a release [here](https://github.com/atombender/go-jsonschema/releases).

* **Install from source**: To install with Go 1.25+:

```shell
go get github.com/atombender/go-jsonschema/...
go install github.com/atombender/go-jsonschema@latest
```

* **Install with Brew**: To install with [Homebrew](https://brew.sh):

```shell
brew tap omissis/go-jsonschema
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
    * [x] `additionalProperties: false` (opt-in via `Config.StrictAdditionalProperties`)
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
    * [x] Email addresses (opt-in via `Config.FormatValidation`)
    * [x] Hostnames (opt-in via `Config.FormatValidation`)
    * [ ] IP addresses
    * [x] Resource identifiers — `uri`, `uri-reference` (opt-in via `Config.FormatValidation`)
    * [ ] URI-template
    * [ ] JSON pointers
    * [x] Regex (opt-in via `Config.FormatValidation`)
    * [x] `uuid` (extension; opt-in via `Config.FormatValidation`)

### Opt-in `format` validation

Setting `Config.FormatValidation.Enabled = true` emits a runtime check on
every `format` keyword listed below. The validator runs after the typed
struct decode, so the field's Go type is already its target form. Optional
pointer fields are skipped when nil, so absent values do not trigger
validation.

| `format` | Strategy | Notes |
| --- | --- | --- |
| `uuid` | RE2 regex against the canonical 8-4-4-4-12 hex form (case-insensitive) | Not in the JSON Schema core spec; supported as a common extension. |
| `email` | `net/mail.ParseAddress` + reject if a display-name is present + require the parsed address to round-trip the input verbatim | Stricter than the Go stdlib default: only RFC 5321 `addr-spec` (`local@domain`) is accepted. Forms like `Alice <alice@example.com>` or `<bob@example.com>` are rejected. |
| `uri` | `net/url.Parse` + `IsAbs()` + RFC 3986 character-class regex | Rejects whitespace, control characters, and malformed percent-encoding (`%` not followed by two hex digits). Empty strings are rejected (not absolute). |
| `uri-reference` | `net/url.Parse` + RFC 3986 character-class regex | Same character-class check as `uri`. Empty strings, fragment-only refs (`#x`), and relative paths (`/a/b`) are accepted per RFC 3986. |
| `hostname` | RE2 regex (RFC 1123 labels) | |
| `regex` | `regexp.Compile` (Go's RE2 syntax) | |

The validation is opt-in to preserve existing behavior. To restrict
validation to a subset of keywords, set `Config.FormatValidation.AllowList`:

```go
cfg.FormatValidation = generator.FormatValidationConfig{
    Enabled:   true,
    AllowList: []string{"uuid", "email"}, // nil = all known formats
}
```

From the CLI, use `--validate-formats`:

```shell
# Validate every supported format keyword.
go-jsonschema --validate-formats=all -p main schema.json

# Validate only specific formats.
go-jsonschema --validate-formats=uuid,email -p main schema.json

# Explicit off (same as omitting the flag).
go-jsonschema --validate-formats=off -p main schema.json
```

Unknown format names are rejected at flag-parse time so a typo like
`uuid,emial` fails immediately rather than silently disabling email
validation.

The `uri` / `uri-reference` validation is **best-effort, not RFC-3986-perfect** —
no Go validator (stdlib or third-party) is — but it catches the common
cases that bare `url.Parse` accepts (whitespace, control chars, malformed
pct-encoding).

### Opt-in `additionalProperties: false` enforcement

`Config.StrictAdditionalProperties` controls whether unknown fields in JSON
or YAML input are rejected at unmarshal time. The three modes are:

| Mode | Behavior |
| --- | --- |
| `""` (off, default) | Silently drop unknown fields. Preserves historical behavior. |
| `"respect-schema"` | Reject unknown fields only for objects whose schema declares `additionalProperties: false`. Other objects continue to drop unknown fields silently. |
| `"strict"` | Reject unknown fields for every generated object type, regardless of what the schema declared. Skipped when the schema specifies a typed `additionalProperties` (a catch-all map field is generated instead). |

Enforcement runs against the raw decoded map, so JSON and YAML inputs are
checked uniformly. `patternProperties` schemas suppress enforcement with a
warning, since the generator has no first-class support for them yet.

```go
cfg.StrictAdditionalProperties = generator.StrictAdditionalPropertiesRespectSchema
```

From the CLI, use `--strict-additional-properties`:

```shell
# Respect the schema's additionalProperties: false declarations.
go-jsonschema --strict-additional-properties=respect-schema -p main schema.json

# Reject unknown fields for every object type.
go-jsonschema --strict-additional-properties=strict -p main schema.json

# Explicit off (same as omitting the flag).
go-jsonschema --strict-additional-properties=off -p main schema.json
```

Unknown mode names (e.g. `--strict-additional-properties=rstrict`) are
rejected at flag-parse time so a typo cannot silently fall back to "off".

## License

MIT license. See `LICENSE` file.
