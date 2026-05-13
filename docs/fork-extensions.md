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
