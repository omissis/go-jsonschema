# `$ref` semantics and design principles

## Background

Historically, `$ref` handling in this project has leaned toward a declaration-first implementation strategy:
referenced schemas are often resolved into reusable generated declarations, aliases, or named types, and
validation is then attached through those generated structures.

That implementation strategy is reasonable from an engineering perspective. It can reduce duplicated generated
types, improve reuse, and avoid unnecessary code expansion.

However, implementation convenience should not define product semantics.

This document exists to clarify the intended semantics of `$ref`, so future changes can be evaluated not only
by whether they work, but also by whether they preserve the intended user-facing model.

---

## Core principle

By default, `$ref` should be **semantically transparent** to schema authors.

In other words, moving a schema fragment:

- inline in place
- into `$defs`
- into another file and referencing it externally

should not, by itself, change the generated behavior.

A `$ref` is, by default, a **reuse and organization mechanism**, not a request to introduce a new semantic
boundary.

---

## Why this matters: examples

### Example 1: extracting a field schema should not change validation

Inline form:

```json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9.-]+$"
    }
  }
}
```

Extracted form:

```json
{
  "type": "object",
  "properties": {
    "name": {
      "$ref": "./name.schema.json"
    }
  }
}
```

With `name.schema.json`:

```json
{
  "type": "string",
  "pattern": "^[a-z0-9.-]+$"
}
```

These two schemas express the same intent. The second form is simply a refactoring of the first into a reusable file.

If extracting the schema changes validation behavior, then `$ref` is not behaving as a transparent reuse mechanism.
That would feel like a programming language where moving code into another file changes what the program does.

This project should avoid that outcome by default.

### Example 2: `format` describes value semantics, not naming intent

Inline form:

```json
{
  "type": "object",
  "properties": {
    "createdAt": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

Extracted form:

```json
{
  "type": "object",
  "properties": {
    "createdAt": {
      "$ref": "./datetime.schema.json"
    }
  }
}
```

With `datetime.schema.json`:

```json
{
  "type": "string",
  "format": "date-time"
}
```

Here the schema author is saying that the field is to be interpreted as a date-time value. They are not, by that fact
alone, requesting a separately modeled semantic concept.

A generator may choose to represent this field as `string`, `time.Time`, or some other Go type. But that backend
representation choice should not redefine the default meaning of `$ref`.

Otherwise, changing internal code generation strategy would also change user-facing schema semantics.

### Example 3: explicit modeling intent is different

```json
{
  "title": "UserName",
  "type": "string",
  "pattern": "^[a-z0-9.-]+$"
}
```

Or:

```json
{
  "type": "string",
  "goJSONSchema": {
    "type": "domain.UserName"
  }
}
```

These cases are different. The schema author is no longer merely extracting a reusable fragment; they are explicitly
asking for a separately modeled concept.

In those cases, preserving a distinct generated symbol boundary is appropriate.

---

## What “transparent” means

By default, transparent `$ref` semantics mean:

- inline schemas and equivalent referenced schemas should behave the same
- internal refs and external refs should behave the same
- validation semantics should not change merely because a schema fragment was extracted
- backend Go type materialization should not redefine the semantic category of a ref
- internal generator optimizations may vary, but user-visible behavior should remain stable

For example, a schema author should be able to extract a field schema into another file without expecting validation
behavior to change.

---

## `$ref` is not automatically a symbol boundary

A referenced schema should not automatically become a separately meaningful generated symbol just because it
was referenced.

A `$ref` becomes a meaningful boundary only when the schema author explicitly expresses that intent.

Examples of explicit modeling intent may include:

- `title`
- `goJSONSchema.type`
- `x-go-type`
- other explicit code-generation extensions that clearly request a named or separately materialized concept

Without such signals, `$ref` should be treated as transparent by default.

---

## Phase-1 local validation overrides for `$ref` siblings

When a schema object contains `$ref`, this project now supports a narrow first phase of local validation
overrides for selected primitive/value-like validation keywords.

Mental model:

1. resolve `$ref`
2. overwrite supported local validation fields from the use-site schema object

This is a direct post-dereference overwrite model. It is intentionally **not** an inferred merge/intersection
model (no "choose stricter value" logic).

### Supported use-site sibling keywords (phase 1)

- String-like: `minLength`, `maxLength`, `pattern`, `const`
- Numeric: `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum`, `multipleOf`, `const`
- Boolean: `const`

### Explicitly out of scope in this phase

Use-site siblings such as `type`, `format`, `title`, object `properties` merging, and implicit object inheritance
are not part of this mechanism.

---

## What should *not* define `$ref` semantics

The following should not, by themselves, decide whether a referenced schema is treated as inline-like or as a
separate semantic symbol:

- whether the ref is internal or external
- whether the ref points into `$defs`
- whether the current backend codegen happens to map the schema to a named Go type
- whether the generator finds it convenient to attach validation at the referenced declaration level
- whether implementation reuse or deduplication is easier in one strategy than another

These are implementation concerns, not author intent.

---

## `format` is value semantics, not naming intent

Keywords such as `format` describe how a value should be interpreted or validated.

For example:

- `format: date-time`
- `format: ipv4`
- `format: email`

These may influence runtime validation or the Go type chosen to represent a value. But they do **not**
automatically mean the schema author intended to create a separately named semantic concept.

A generator may choose to represent `date-time` as `time.Time`, or may choose a different internal
representation in the future. That implementation choice should not change whether `$ref` is considered
transparent.

Otherwise, product semantics would become unstable under backend implementation changes.

---

## Objects and internal optimization

Object schemas are different from primitive/value-like schemas in one important respect: in Go, they naturally
tend to materialize as separate struct types.

That is acceptable.

The generator may optimize object handling by:

- reusing declarations
- deduplicating generated structs
- importing generated types across packages
- avoiding unnecessary expansion

But these are still implementation optimizations. They should not change the semantic expectation that
extracting an object schema behind a `$ref` should preserve behavior unless the author explicitly requested a
distinct modeling boundary.

The local override mechanism above applies only to a narrow whitelist of primitive/value-like validation fields.
It should not be interpreted as general-purpose object extension or inheritance.

---

## Validation override vs object extension/composition

Local `$ref` validation override and object composition are separate concerns:

- **Local validation override**: overwrite selected primitive/value-like validation fields at the use-site after
  dereference.
- **Object extension/composition**: explicitly combine object schemas (for example, with `allOf`) when structural
  extension is intended.

If you want to extend an object schema structurally, treat that as an explicit composition problem rather than
implicit sibling-property merging on a `$ref` object.

---

## Unnamed vs explicitly-named object refs

A subtle but important distinction applies to object schemas referenced from `$defs`:

**Unnamed object ref** (no `title`, `goJSONSchema.type`, or `x-go-type`):

```json
{
  "$defs": {
    "SubTypeB": {
      "type": "object",
      "properties": { "x": { "type": "string" } }
    }
  }
}
```

Multiple references to such a schema may materialize contextually. The `$defs` key name alone does not
constitute explicit modeling intent. The generator may reuse a generated struct by name (as an implementation
optimization), but it is not required to treat the defs key as a stable user-visible symbol.

**Explicitly named object ref** (with `title` or other explicit extension):

```json
{
  "$defs": {
    "SubTypeC": {
      "type": "object",
      "title": "SubTypeC",
      "properties": { "x": { "type": "string" } }
    }
  }
}
```

Here the schema author has signaled a distinct modeling concept. The generator should preserve a stable
shared symbol across all references to this schema.

**In practice**, the distinction matters most when the same unnamed object fragment is referenced from
multiple fields. For example:

```json
"b1": { "$ref": "./defs.schema#/$defs/SubTypeB" },
"b2": { "$ref": "./defs.schema#/$defs/SubTypeB" }
```

vs.

```json
"c1": { "$ref": "./defs.schema#/$defs/SubTypeC" },
"c2": { "$ref": "./defs.schema#/$defs/SubTypeC" }
```

For `SubTypeB` (unnamed): sharing a generated struct is an acceptable optimization, not a semantic requirement.
For `SubTypeC` (explicitly titled): all references should resolve to the same generated symbol.

---

## Separation of concerns for implementation

Future implementation work should try to keep the following concerns separate:

1. **`$ref` semantic policy**
   - Should this reference be treated transparently?
   - Has the author explicitly requested a separate modeling boundary?

2. **Go type materialization**
   - Should the value be represented as `string`, `time.Time`, `netip.Addr`, a custom type, etc.?

3. **Validator attachment**
   - Should validation happen at the field level, the referenced declaration level, or both?

These concerns interact, but they should not be collapsed into one decision.

In particular, backend type materialization should not implicitly decide semantic `$ref` behavior.

---

## Non-goals

This document does **not** require that:

- all refs must be mechanically text-inlined
- all referenced objects must be flattened
- all declaration reuse must be removed
- all implementation strategies must look the same internally

This document only defines the intended product semantics and the architectural boundaries that future changes
should respect.

---

## Practical guidance for future contributors

When changing `$ref` behavior, ask:

1. Does this change alter behavior only because a schema moved behind a ref?
2. Does this change make internal vs external refs behave differently?
3. Does this change let backend Go type mapping influence semantic `$ref` policy?
4. Is a separate symbol being preserved because the author asked for it, or only because the implementation
   happened to produce one?

If the answer depends mainly on implementation convenience rather than schema author intent, the design should
be reconsidered.

---

## Summary

The intended direction is:

- `$ref` is transparent by default
- extraction across files or `$defs` should not, by itself, change semantics
- explicit modeling intent may preserve a separate symbol boundary
- internal generator optimizations are allowed
- implementation details must not redefine user-facing schema meaning
