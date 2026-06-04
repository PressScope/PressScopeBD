# Analytics Service

## Commands

- **Test**: `go test ./...` (runs tests for all packages)
- **Build**: `go build -o /tmp/analytics-server ./cmd/server && go build -o /tmp/analytics-worker ./cmd/motherduck`
- **Lint**: `golangci-lint run ./...` (if golangci-lint is installed)

## Environment Files

- `.env` - default config (production mode, no migrations)
- `.env.development` - development mode with local valkey, auto-creates tables
- `.env.production` - production mode template (token placeholder)

## Migration Behavior

- Auto-creates events table only when `APP_ENV=development`
- Production requires pre-existing table structure