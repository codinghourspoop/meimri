# memri

A Svelte + Go flashcard app. The Go server owns auth, sessions, and data; the
Svelte SPA is a static bundle served by it.

## Requirements

Everything comes from the flake — run `direnv allow` once and the shell picks up
the toolchain (Go, Task, goose, Bun, Node, dotenvx):

```sh
direnv allow
```

## Configuration

Copy `.env.example` to `.env` and fill in the values. `.env` is gitignored and
must never be committed.

```sh
cp .env.example .env
```

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `ADDR` | no | `:8080` | Listen address |
| `DATABASE_URL` | no | `./data/app.db` | SQLite path, a Turso URL, or `:memory:` |

Env files are loaded by **dotenvx**, not by the Go process and not by direnv.
Anything that needs the values is wrapped in `dotenvx run --` inside the
Taskfile:

```sh
task dev          # Vite + Go API
task dev:api      # Go API only
task migrate      # runs against DATABASE_URL from .env
```

If you run the binary directly (`go run ./cmd/server`), it reads
`os.Getenv` only — export the values yourself or prefix with
`dotenvx run --`.

## Development

```sh
task dev            # Vite dev server + Go API via air (parallel)
task dev:api        # Go API only; Vite proxies /api to :8080
task dev:frontend   # Vite only
```

## Database

The server runs pending migrations on boot, so a fresh checkout needs no
manual step. `task migrate` and `task migrate:down` are for when you want to
control timing.

Migrations live in `db/migrations` and use the SQLite dialect.

## Build

```sh
task build          # Vite production bundle + Go binary
```

`task build` writes the SPA to `static/dist/` and the binary to `./tmp/main`.
The server serves `static/dist` with an SPA fallback, so build before
`./tmp/main` if you are not using the Vite dev server.

## Tests

There is no test suite yet. `go build ./...` and `go vet ./...` are the
available checks.
