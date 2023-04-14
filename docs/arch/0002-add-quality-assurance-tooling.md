# 2. Add Quality Assurance tooling

Date: 2023-04-09

## Status

Accepted

## Context

We want to increase the internal quality of the project.

## Decision

We are going to introduce a number of tools and practices to help with most aspects of development, including:

- Writing down Architectural Decision Records
- Ensuring all required tooling is installable with a few commands
- Running static code analysis wherever possible on all files
- Developing automated tests to cover most of the codebase
- Setting up CI pipelines to ensure all checks are run all the time
- Keeping a consistent format when redacting commit messages and release notes

In general, we want to automate all aspect of the project development lifecycle and leverage tools to the maximum extent to ensure we can focus our attention on delivering value

## Consequences

The project should reach a high level of automation and quality, guaranteed by practices and tools, while keeping maintenance costs at a reasonable level.
