

# gocamelpack 🐫

A Go-powered CLI that ingests photos and other media, tags them via EXIF,
and copies them into a date-based hierarchy on disk—your lightweight photo
off‑loader without the heavy UI.

---

## Features

* **EXIF‑aware copying** – destination paths are built from `CreationDate`
  metadata (`YYYY/MM/DD/HH_mm.ext`).
* **Safe by default** – never overwrites unless you pass `--overwrite`.
* **Dry‑run mode** – preview every copy before bytes move.
* **Pluggable concurrency** – upcoming `--jobs` flag will parallelise copies.
* **Idiomatic Go API** – all logic lives under `files/`, easy to import.

---

## Quick start

```bash
# build (installs into $GOPATH/bin or $GOBIN)
go install ./...

# copy a single file
gocamelpack copy ~/Downloads/IMG_1234.JPG /Volumes/Photos

# copy everything in a folder, but preview first
gocamelpack copy --dry-run ~/Downloads/DCIM /Volumes/Photos
```

---

## Command flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--dry-run`   | `false` | Print planned copies without executing them. |
| `--overwrite` | `false` | Allow clobbering destination files. |
| `--jobs`      | `1`     | Worker count for concurrent copies (coming soon). |

---

## Development

```bash
# run linters & tests
go vet ./...
go test ./...
```

### Code coverage

```bash
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out
```

---

## Project structure

```
cmd/      - Cobra CLI commands
files/    - Core file‑handling logic (EXIF, copy, validation)
deps/     - Thin wiring between CLI and services
```

---

## Roadmap / TODOs

- [ ] A solid testing setup to be confident whats going to happen to the files.
- [ ] An option for all or nothing copy, with graceful failure and cleanup.
- [ ] Progress indicator during copies.
- [ ] Signal (Ctrl‑C) cancellation with cleanup.
- [ ] Optional checksum verification (`--verify-sha256`).
- [ ] User defined output folder structure.
    - Instead of basing it off creation date for my archival purposes, a user should be able to define the the output path. 
    - It should be able to use the exif data as a resource.
    - There should be some defined syntax, probably just object notation.
- [ ] implement worker pool behind `--jobs`.
- [ ] Extended attributes & timestamp preservation flag.
- [ ] CI: add GitHub Actions for `go vet`, tests, and coverage gate.
- [ ] Documentation polish, examples with screenshots / asciinema.

---

## License

MIT