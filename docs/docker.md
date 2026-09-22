# Container-based setup

The repository's build, verification, and CLI runtime run in a Podman container.
The host does not need Node.js, npm, pnpm, or project dependencies.

## Prerequisites

Install Podman and ensure the current user can run `podman`.

## Build and run

```sh
make container-build
make container-run
```

`make install-mcode` is retained as a compatibility target; it builds the local
image and does not install a host-side Node package. `make-q1` runs `pnpm verify`
inside the image with network access disabled:

```sh
make make-q1
```

The image sets `MCODE_DISABLE_TELEMETRY=1` and `DO_NOT_TRACK=1` by default. No
host credentials, sessions, environment files, or user configuration are copied
into the image. Provider credentials must be passed explicitly at runtime when
needed, using a narrowly scoped environment or mounted configuration directory.

## Boundary status

This is the first container milestone. A future companion CLI is still required
to make Podman the validated host boundary described in issues 002 and 003; the
Make targets currently invoke Podman directly and should not be treated as that
allowlisted Go boundary.
