# Yummi

A self-hosted recipe manager that prioritizes simplicity over feature creep.

## Why Yummi?

Most recipe managers try to be everything: meal planners, grocery list generators, nutrition calculators, social networks. Yummi does one thing well: store, organize, and find your recipes.

### Single binary, zero dependencies

```bash
./yummi
```

That's it. No Docker compose files, no PostgreSQL, no Redis, no Node.js build pipeline, no nginx reverse proxy configuration. One binary, one SQLite file. Put it on any Linux box and it runs.

### AI-assisted import that actually works

Paste a URL from any recipe site. Yummi extracts the recipe via Schema.org JSON-LD, then optionally:

- **Cleans up instructions** — removes garbage steps like section headers ("Zubereitung") that parsers wrongly include as steps
- **Rephrases for consistency** — condenses 28 repetitive layering steps into one clear sentence, standardizes voice and formatting across all your recipes
- **Classifies automatically** — suggests matching categories and tags from your existing taxonomy

Each of these is independently toggleable. No AI key? The parser works fine without it.

### Native Android app

Not a PWA in a webview. A proper Kotlin/Jetpack Compose app with Material 3, pull-to-refresh, haptic feedback, and offline-friendly image caching. Talks to the same REST API as the web UI.

### Designed for humans, not robots

- Full-text search with prefix matching — type "bor" and find "Borschtsch"
- Recipe sharing between users — no complex permission systems, just "share with"
- "Tried" tracking — mark what you've actually cooked
- Categories and tags — click them anywhere to filter
- Light/dark theme — persisted per device

### What Yummi deliberately doesn't do

- Meal planning
- Grocery lists
- Nutrition calculation
- Social features / public profiles
- Plugin systems
- Multiple databases
- Kubernetes deployment manifests

## Tech stack

| Component | Choice | Why |
|-----------|--------|-----|
| Backend | Go + `net/http` | Single binary, no framework churn |
| Database | SQLite (pure Go, no CGO) | Zero ops, embedded, fast |
| Web UI | Go templates + HTMX + Pico CSS | No JS build step, instant page loads |
| Android | Kotlin + Jetpack Compose | Native Material 3 experience |
| Search | SQLite FTS5 | Full-text search without Elasticsearch |
| AI | OpenAI API (optional) | gpt-4.1-nano/mini for classification and cleanup |

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

## Environment variables

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
