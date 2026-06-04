# Fork extensions

## `x-go-oneof-envelope`

`x-go-oneof-envelope` is a fork-specific extension for discriminator-driven
envelope shapes such as:

```json
{ "type": "...", "value": { ... } }
```

It is intended for explicit envelope routing, not generic `oneOf` inference.

### Supported patterns

- Single envelope field on an object (for example `type -> value`)
- Multiple envelope fields on the same object (for example `type -> value` and
  `mode -> payload`)
- Nested envelope fields (for example `nested_type -> nested_value`)

### Extension shape

Apply `x-go-oneof-envelope` to a payload field that contains `oneOf`:

```json
{
  "x-go-oneof-envelope": {
    "discriminator": "type",
    "mapping": {
      "a": "AValue",
      "b": "BValue"
    }
  }
}
```

- `discriminator` is the sibling field used to select the branch.
- `mapping` keys are discriminator values.
- `mapping` values must match the resolved `title` of a `oneOf` branch.

### Branch titles and mapping rules

- `oneOf` branches may be inline or `$ref`.
- Each branch must resolve to a unique `title`.
- Mapping targets that title, not the Go type name directly.

### Scope and non-goals

- This extension enables deterministic envelope routing based on
  discriminator/mapping pairs.
- It does **not** add generic automatic branch inference for arbitrary `oneOf`
  unions without envelope metadata.

### Example schema

```json
{
  "type": "object",
  "required": ["type", "value"],
  "properties": {
    "type": { "type": "string", "enum": ["a", "b"] },
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

### Generated routing shape

```go
switch result.Mode {
case MultiOneOfEnvelopeModeX:
	result.Payload = ModePayload{X: &v}
case MultiOneOfEnvelopeModeY:
	result.Payload = ModePayload{Y: &v}
case MultiOneOfEnvelopeModeZ:
	result.Payload = ModePayload{Z: &v}
}

switch result.Type {
case MultiOneOfEnvelopeTypeA:
	result.Value = ValuePayload{A: &v}
case MultiOneOfEnvelopeTypeB:
	result.Value = ValuePayload{B: &v}
case MultiOneOfEnvelopeTypeC:
	result.Value = ValuePayload{C: &v}
}
```

## `x-go-ref`

`x-go-ref` lets a referenced schema reuse an existing Go type from another
package instead of generating that type again.

This is useful when multiple schemas share a common model that already exists in
Go code and you want generated fields to point to that external type directly.

### Extension shape

Apply `x-go-ref` on a schema definition that will be referenced by `$ref`:

```json
{
  "x-go-ref": {
    "path": "github.com/example/shared",
    "alias": "shared"
  }
}
```

- `path` is the Go import path for the existing package.
- `alias` is the import alias to use in generated code.
- `alias` must be a valid Go identifier.
- `path` and `alias` are both required when `x-go-ref` is present.

### How it works

When another schema references a definition marked with `x-go-ref`, generated
code uses `alias.TypeName` for that field and adds the corresponding import,
instead of emitting a new local type declaration for the referenced definition.

The Go type name is resolved from:

- the schema `title`, when `StructNameFromTitle` is enabled and `title` is set;
  otherwise
- the referenced definition name.

`x-go-ref` affects referenced definitions. If the root schema itself has
`x-go-ref`, the generator still emits the local root declaration rather than
replacing it with an external import.

### Example: shared definition imported by reference

Shared schema:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/refXGoImport/sharedTypes",
  "$defs": {
    "User": {
      "type": "object",
      "x-go-ref": {
        "path": "github.com/example/shared",
        "alias": "shared"
      },
      "properties": {
        "id": {
          "type": "string"
        }
      }
    }
  }
}
```

Consumer schema:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "id": "https://example.com/consumer/def-ref",
  "type": "object",
  "properties": {
    "user": {
      "$ref": "./shared.schema#/$defs/User"
    }
  }
}
```

Generated field shape:

```go
type Consumer struct {
	User *shared.User `json:"user,omitempty"`
}
```

The generated file imports `github.com/example/shared` with alias `shared` and
does not generate a local `type User struct`.

### Example: using `title` as the external type name

If `StructNameFromTitle` is enabled and the referenced definition has a `title`,
that title determines the imported type name:

```json
{
  "$defs": {
    "SharedEntity": {
      "title": "User",
      "type": "object",
      "x-go-ref": {
        "path": "github.com/example/shared",
        "alias": "shared"
      }
    }
  }
}
```
``shared.User`` will be used for references to that definition.
