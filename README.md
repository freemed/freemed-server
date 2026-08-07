# FREEMED SERVER

[![Build Status](https://github.com/freemed/freemed-server/actions/workflows/go.yml/badge.svg)](https://github.com/freemed/freemed-server/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/freemed/freemed-server)](https://goreportcard.com/report/github.com/freemed/freemed-server)
[![codecov](https://codecov.io/gh/freemed/freemed-server/branch/master/graph/badge.svg)](https://codecov.io/gh/freemed/freemed-server)
[![GoDoc](https://godoc.org/github.com/freemed/freemed-server?status.png)](https://godoc.org/github.com/freemed/freemed-server)

Refactoring of **FreeMED** in Golang / Gin + SvelteKit.

## Backend

| Component | Technology |
|-----------|-----------|
| Language | Go 1.26 |
| Web framework | Gin |
| Auth | gin-jwt v2 (bcrypt, JWT with jti blacklist via Redis) |
| Database | MySQL via go-sql-driver/mysql |
| Query layer | sqlc (type-safe SQL generation, ~100 query files) |
| Migrations | golang-migrate (17 migrations) |
| Sessions | Redis via go-redis |
| Logging | lumberjack (rolling logger) |
| Graceful shutdown | manners |

## Frontend

| Component | Technology |
|-----------|-----------|
| Framework | SvelteKit 5 (runes mode) |
| Styling | Tailwind CSS 4 |
| Calendar | FullCalendar 6 |
| Adapter | @sveltejs/adapter-static (SPA fallback) |
| Build | Vite |

## Quick Start

```bash
# Backend
make binary && ./freemed

# Frontend dev server
make frontend-deps frontend-dev

# Full stack with Docker
make docker-build docker-up
```

## API Coverage

127+ API endpoints across 45 modules. See `.hermes/plans/` for feature parity analysis.

### Core
`config`, `dashboard`, `search`, `users`, `facilities`, `providers`, `acl`, `tools`

### Patient
`patients`, `patient` (info, addresses, tags, history, diagnoses, photo-id, phones, allergies, vitals, medications, immunizations, coverages)

### Clinical
`encounters`, `progress-notes`, `labs`, `drug-samples`, `episodes-of-care`, `growth-charts`, `workflow-status`, `clinical-orders`

### Scheduler
`scheduler` (appointments, recurring, next-available, blocks, groups, templates)

### Billing
`procedures`, `payments` (ledger, copays, deductibles), `claims`, `authorizations`, `aging`, `remitt`, `superbills`, `action-items`, `superbill-templates`

### Messaging
`messages` (tags, bulk ops), `notifications` (task inbox, timestamps)

### Documents
`documents` (unfiled, unread), `scanned-documents`

### Other
`annotations`, `letters`, `correspondence`, `referrals`, `events`, `holidays`, `callin`, `reporting`, `rx-refill`, `sms-providers`, `provider-specialties`, `emr/data-store`, `support` (picklists)

## Architectural Changes from FreeMED 0.9.x

- **sqlc + database/sql** replaces GORM (compile-time SQL validation)
- **bcrypt** password hashing with automatic MD5 legacy upgrade
- **JWT token blacklist** (Redis-backed jti) — logout invalidates tokens
- **RBAC middleware** (`RequireRole`) replaces hardcoded ACL checks
- **SvelteKit 5 SPA** replaces GWT + jQuery/Knockout/Bootstrap frontend
- **golang-migrate** replaces GORM AutoMigrate for versioned schema changes
- **Redis sessions** replace MySQL-backed sessions
- Environment variable config overrides (`FREEMED_DB_HOST`, etc.) for Docker

## Caveats

- MySQL's `ONLY_FULL_GROUP_BY` needs to be disabled for legacy queries
- `billing/` module requires external `remitt-server` and `ratago` dependencies
- Code compatible with FreeMED 0.9.x database schemas

## Other Resources

- Background image: [CC BY-SA 2.0](https://commons.wikimedia.org/wiki/File:Laptop_and_stethoscope_(6123892769).jpg)
