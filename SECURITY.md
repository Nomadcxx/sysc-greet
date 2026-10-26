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

This repository has no scan of its own tree yet. The scan in the shell repository covers the code

this repository is pinned into, because that is where the module graph is resolved: `govulncheck`

reports the published flaws whose symbols the calling module reaches, and the calling module here is

the shell. The scan of this tree takes the shape of the shell's: a weekly cron beside the release

workflow, with the scanner version pinned in it.

A green scan is not a warranty that the tree has no flaw in it. The scan reads the advisories the Go

project publishes; a flaw in this repository's own code is not in that feed.

## What is out of scope

- A flaw in a dependency, which belongs to the dependency.
- A crash with no security consequence.
- A missing feature.
