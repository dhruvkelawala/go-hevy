---
name: hevy-cli
description: "Hevy workout tracking CLI. List workouts, exercises, routines, track progression, analyze training splits, detect plateaus, check WHOOP readiness, export data. Use when user asks about workouts, gym sessions, exercise history, fitness progress, or Hevy data."
---

# hevy-cli — Hevy Workout Tracker CLI

Repo: https://github.com/dhruvkelawala/hevy-cli

## Installation

### Option 1: GitHub Release (preferred for agents)
```bash
# Detect OS and arch, download latest release
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH="amd64"; [ "$ARCH" = "aarch64" ] && ARCH="arm64"
LATEST=$(curl -sL https://api.github.com/repos/dhruvkelawala/hevy-cli/releases/latest | grep tag_name | cut -d'"' -f4)
curl -sL "https://github.com/dhruvkelawala/hevy-cli/releases/download/${LATEST}/hevy_${OS}_${ARCH}.tar.gz" | tar xz -C /tmp
mv /tmp/hevy ~/.local/bin/hevy
chmod +x ~/.local/bin/hevy
```

### Option 2: Go install
```bash
go install github.com/dhruvkelawala/hevy-cli@latest
```

### Option 3: Build from source
```bash
git clone https://github.com/dhruvkelawala/hevy-cli.git
cd hevy-cli
go build -o ~/.local/bin/hevy .
```

## Authentication

After installing, the user needs a Hevy API key.

```bash
# Interactive setup
hevy init

# Or set directly
export GO_HEVY_API_KEY="<key>"

# Or store in config
hevy config set api_key "<key>"
```

For agent sessions: `source ~/.config/openclaw/env.sh` (if key is stored there).

## Quick Reference

### Account
```bash
hevy status        # Verify API access
hevy me            # User profile
hevy count         # Total workout count
```

### Workouts
```bash
hevy last          # Most recent workout
hevy today         # Today's workout
hevy workouts      # Recent workouts
hevy workout <id>  # Workout detail
```

### Exercises & Routines
```bash
hevy exercises --search "bench press"
hevy exercise <id>
hevy history <exercise-id>
hevy routines
hevy routine <id>
```

### Training analytics
```bash
hevy progress "Squat"   # ASCII progression chart
hevy streak             # Weekly streak
hevy pr --all           # All personal records
hevy week               # Weekly summary
hevy diff               # Compare last two workouts
hevy volume "Squat"     # Volume over time
hevy muscles            # Muscle groups hit this week
hevy calendar           # ASCII workout calendar
hevy search upper       # Search workouts
```

### Advanced insights
```bash
hevy plan               # Suggest next workout
hevy consistency        # Training consistency report
hevy plateau            # Detect stalled exercises
hevy supersets          # Superset pairings
hevy fatigue            # RPE trend analysis
hevy split              # Actual training split
hevy records            # All-time vs current bests
hevy rest               # Time efficiency
hevy readiness          # WHOOP recovery + training advice
```

### Export
```bash
hevy export --format csv
```

## Common Agent Tasks

| User asks | Command |
|---|---|
| "What did I do at the gym?" | `hevy last` |
| "How's my squat progressing?" | `hevy progress "Squat"` |
| "Show my PRs" | `hevy pr --all` |
| "What should I train next?" | `hevy plan` |
| "Am I ready to train today?" | `hevy readiness` |
| "How consistent am I?" | `hevy consistency` |
| "Am I plateauing?" | `hevy plateau` |
| "Export my data" | `hevy export --format csv` |

## Output Flags
- Default: formatted table
- `--json` / `-j`: JSON for scripting
- `--compact`: one line per item
- `--kg` / `--lbs`: unit toggle (default kg)

## WHOOP Integration (optional)
```bash
hevy config set whoop_path /path/to/whoop-tracker-skill
hevy readiness
```

## Notes
- API requires Hevy Pro subscription for full access
- Exercise search paginates all pages (may take 1-2s)
- Progress does fuzzy name matching — use exact name for best results
