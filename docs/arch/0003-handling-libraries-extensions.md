# 3. Handling libraries extensions

Date: 2023-04-09

## Status

Accepted

## Context

We sometimes find ourselves in a situation where we need to extend a library that we use. This can be for a number of reasons, but the most common ones are that we need to add a feature that the library doesn't support, or to add one or more helper functions that we need to make our life easier, or to reduce duplication, or to refactor away some non-business logic from our codebase to the library.

## Decision

We will store all extensions and helpers to third-party and golang own's libraries with the `x/` folder. For example, helper functions for the `net/http` package will be stored in `x/net/http`.

## Consequences

This will allow us to keep our codebase clean and easier to read, and it will keep the folders containing the business logic more focused on the tasks to solve.
