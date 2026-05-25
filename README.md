# Yummi

A self-hosted recipe manager. Single binary, SQLite, German UI.

## Overview

Yummi stores, organizes, and searches recipes. It runs as a single Go binary with an embedded SQLite database — no external services required.

## Features

- **URL import** — paste a link from any recipe site, extracts via Schema.org JSON-LD
- **Full-text search** — SQLite FTS5 with prefix matching
- **Categories and tags** — filterable, clickable everywhere
- **Recipe sharing** — share recipes with other users on the same instance
- **"Tried" tracking** — mark recipes you've cooked
- **Light/dark theme** — persisted per device
- **Optional AI processing** — on import, can clean up instructions, rephrase for consistency, and suggest categories/tags (each independently toggleable)
- **Native Android app** — Kotlin/Jetpack Compose, Material 3, connects via REST API
- **Admin panel** — user management, settings, data export/import

## Tech stack

| Component | Choice |
|-----------|--------|
| Backend | Go, `net/http`, single binary |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| Web UI | Go templates, HTMX, Pico CSS |
| Search | SQLite FTS5 |
| Android | Kotlin, Jetpack Compose, Material 3 |
| AI | OpenAI API (optional, gpt-4.1-nano/mini) |

## Quick start

```bash
# Build
go build -o yummi .

# Run (creates ./data/yummi.db on first start)
./yummi

# With options
./yummi -port :9090 -data /srv/yummi/data

# With AI features
YUMMI_OPENAI_KEY=sk-... ./yummi

# Create initial admin
./yummi -create-admin "admin:password"
```

Visit `http://localhost:8080`, register a user, start importing recipes.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `YUMMI_PORT` | `:8080` | Listen address |
| `YUMMI_DATA_DIR` | `./data` | Database and uploads directory |
| `YUMMI_SECRET` | (random) | Session signing secret |
| `YUMMI_OPENAI_KEY` | (empty) | OpenAI API key for AI features |

## Android app

The Android app lives in `./android/` and connects to any Yummi server.

```bash
cd android
JAVA_HOME=/usr/lib/jvm/java-21-openjdk ./gradlew assembleDebug
adb install app/build/outputs/apk/debug/app-debug.apk
```

## License

MIT
