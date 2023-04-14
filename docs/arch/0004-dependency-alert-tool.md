# 4. Dependency alert tool

Date: 2023-04-09

## Status

Accepted

## Context

We need a tool that helps us update the project dependencies automatically, keeping it as secure as possible.

## Decision

We are going to use [Renovate](https://github.com/renovatebot/renovate) as the dependency maintainer: it is a third-party software installable as Github App that simplifies the update process, opening a PR when a new dependency update is available. I prefer this over Dependabot because of its flexibility and more advanced features.
See [this article](https://javascript.plainenglish.io/automate-dependency-updates-by-renovate-not-by-dependabot-6efddd549a3e) for a more detailed comparison.

## Consequences

We will have to spend less time updating dependencies manually, and the project will be more secure.
