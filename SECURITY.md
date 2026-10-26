# Security policy

## Reporting a vulnerability

Report it privately, through GitHub's private advisory form for this repository:

<https://github.com/Nomadcxx/sysc-greet/security/advisories/new>
Use the private form rather than a public issue. A public issue tells every reader of the repository
that the flaw is there before it is fixed.

Write what an attacker can do, how to do it, and the version or the commit the flaw is in. A proof
of concept is worth more than a description of the flaw.

## Which versions are supported

| Version | Supported |
|---|---|
| The latest tag | Yes |
| The tag before it | Yes |
| Anything older | No |

`sysc-greet` is released as tags on `master`. A flaw is fixed on `master` and the fix reaches a user
in the next release.

## What a report gets

| Stage | Time |
|---|---|
| Acknowledgement | three working days |
| Assessment | ten working days |
| Fix | depends on the severity |

## The standing scan

This repository does not scan its own tree. The `govulncheck` scan in `sysc-shell` checks the
shell's module graph and reports published vulnerabilities in reachable symbols. This repository is
not a dependency of `sysc-shell`, so the shell scan does not cover it.

A green shell scan says nothing about vulnerabilities in this repository's own code. `govulncheck`
reports advisories published by the Go project; vulnerabilities outside that feed are not reported.

## What is out of scope

- A flaw in a dependency, which belongs to the dependency.
- A crash with no security consequence.
- A missing feature.
