# go-hevy

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`go-hevy` is a Go CLI for the [Hevy](https://www.hevyapp.com) API. It gives you scriptable terminal access to workouts, routines, exercise templates, history, and profile data.

## Installation

### Go install

```bash
go install github.com/dhruvkelawala/go-hevy@latest
```

### Homebrew / binaries

Homebrew and release binaries are published from GitHub Releases.

## Quick start

```bash
# Interactive setup
hevy init

# Recent workouts
hevy workouts

# Workout details
hevy workout <id>

# User profile
hevy me
```

## Authentication

Hevy uses an `api-key` header.

You can configure it in either place:

- `~/.config/go-hevy/config.json`
- `GO_HEVY_API_KEY` environment variable

Environment variables take precedence over the config file.

## Commands

### Core

| Command | Description |
| --- | --- |
| `hevy workouts` | List recent workouts |
| `hevy workouts --json` | List workouts as JSON |
| `hevy workout <id>` | Show workout details |
| `hevy workout create -f workout.json` | Create a workout from JSON |
| `hevy workout update <id> -f workout.json` | Update a workout from JSON |
| `hevy count` | Show total workout count |
| `hevy routines` | List routines |
| `hevy routine <id>` | Show routine details |
| `hevy exercises` | List exercise templates |
| `hevy exercise <id>` | Show exercise template details |
| `hevy history <exercise-id>` | Show exercise history |
| `hevy me` | Show user info |

### Utility

| Command | Description |
| --- | --- |
| `hevy init` | Interactive API key setup |
| `hevy config` | Show current config |
| `hevy config set key <value>` | Store API key in config |
| `hevy version` | Print version |

## Output formats

All list/detail commands support:

- table output (default)
- JSON via `--json` / `-j`
- compact via `--compact`

Examples:

```bash
hevy workouts
hevy workouts --json
hevy workout <id> --compact
```

## Pagination

List commands support:

```bash
hevy workouts --page 2 --page-size 5
hevy exercises --page-size 25
```

Page size limits:

- workouts: max 10
- routines: max 10
- exercises: max 100

## Workout create/update JSON shape

```json
{
  "workout": {
    "title": "Friday Leg Day 🔥",
    "description": "Medium intensity leg day focusing on quads.",
    "start_time": "2024-08-14T12:00:00Z",
    "end_time": "2024-08-14T12:45:00Z",
    "is_private": false,
    "exercises": [
      {
        "exercise_template_id": "D04AC939",
        "notes": "Felt strong today.",
        "sets": [
          {
            "type": "normal",
            "weight_kg": 100,
            "reps": 10,
            "rpe": 8.5
          }
        ]
      }
    ]
  }
}
```

## AI agent integration example

```bash
hevy workouts --json | jq '.workouts[] | {id, title, start_time}'
```

## Development

```bash
go build ./...
go test ./...
```

## Release

GitHub Actions builds tagged releases for macOS, Linux, and Windows using GoReleaser.

## License

MIT
