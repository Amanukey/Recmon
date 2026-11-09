# ReconFlow

Local-first reconnaissance orchestration platform. This repository starts with the M0/M2 foundation: a Go executable bound to `127.0.0.1`, token-protected API routes, and an embedded dashboard shell.

## Run

```bash
go run ./cmd/reconflow
```

ReconFlow prints a local URL and random access token at launch. The initial dashboard is intentionally read-only while the engine and SQLite writer are built out.

## Safety

Only use ReconFlow against assets you are authorized to assess. Localhost binding, host validation, and a launch token are defense-in-depth; they are not a substitute for OS-level protection.
