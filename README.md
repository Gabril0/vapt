# vapt

A terminal typing test with online features — leaderboards, tenants (organizations), and profile management.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

`vapt` is the **terminal client**. Online features talk to a **`vupt` server** backend
over HTTP. By default it points at a hosted instance, but you can run your own server
and point the client at it (see [Server](#server)).

## Features

- **Typing test** — timed tests (15s / 30s / 60s / 120s) with live WPM, accuracy, and consistency tracking
- **Offline-first** — works without an account; best score saved locally
- **Leaderboards** — filter by metric (score / WPM / accuracy / consistency), duration, and period (defaults to daily score)
- **Tenants** — create or join organizations; tenant-scoped leaderboards and results
- **My Results** — paginated history with sorting and filtering
- **Profiles** — view/edit username and password
- **Invite codes** — admin-managed access control for tenants
- **Themes** — customizable color palette via `theme.json`
- **Custom words** — drop a `words.txt` file to use your own word list

## Install

Build from source:

```
go build -o vapt ./cmd
```

## Usage

```
./vapt
./vapt -version   # print version and exit
```

Navigate with keyboard shortcuts shown on each screen. Press `esc` to go back.

The typing test and local best score work fully **offline**. Only the online
features (leaderboards, tenants, profiles) require a reachable `vupt` server.

## Server

`vapt` speaks to a `vupt`-compatible HTTP backend. Select which server to use with
the `VUPT_URL` environment variable:

```
VUPT_URL=http://localhost:8080 ./vapt
```

If `VUPT_URL` is unset, a default hosted instance is used. You can implement and
self-host your own server — the client only relies on the REST endpoints defined
in `internal/api/client.go` (auth, tenants, leaderboards, results, invite codes).

## Configuration

### Theme

Create a `theme.json` in the working directory:

```json
{
  "accent": "#ff79c6",
  "text": "white",
  "correct": "green",
  "error": "red"
}
```

Supports named colors, hex (`#RRGGBB`), `rgb(R,G,B)`, and raw ANSI codes.

### Custom words

Create a `words.txt` file (one word per line) to replace the default text generator.

### Local data

The token, active tenant, and local best score are stored under `~/.vapt/`.

## Project structure

```
cmd/main.go               Entry point (version flag, program bootstrap)
internal/
  app/                    Core TUI application
    model.go              State, model struct, helpers
    update.go             Bubble Tea Update handler
    views.go              Bubble Tea View renderer
    online.go             Async message types and commands
  api/
    types.go              API request/response types
    client.go             HTTP client for the vupt backend
  auth/auth.go            Token and tenant persistence (~/.vapt/)
  storage/storage.go      Local best-score storage (~/.vapt/best.dat)
  ui/
    style.go              ANSI styling and box-drawing utilities
    theme.go              Theme loading and color parsing
  words/words.go          Text generation (custom words + gofakeit)
```

## License

MIT
