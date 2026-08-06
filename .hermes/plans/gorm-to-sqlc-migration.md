# GORM → sqlc Migration Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Replace GORM ORM with sqlc for type-safe, compile-time-checked SQL query
generation, eliminating the runtime ORM layer while keeping the existing MySQL database
and Gin web framework.

**Architecture:** GORM is already used mostly as a raw SQL passthrough (23/29 calls are
`Db.Raw()`). Only 6 calls use GORM's ORM features (Find, First, Save). The migration
replaces GORM's connection pool with `database/sql` + `go-sql-driver/mysql`, generates
sqlc query functions for the 6 ORM-style calls, converts 23 raw-SQL queries to sqlc
`:many`/`:one` named queries, and removes the `gorm.Model` embedding from 53 structs.

**Tech Stack:** Go 1.26, MySQL (go-sql-driver/mysql), sqlc v1.28+, Gin, existing Redis
sessions, existing JWT auth (gin-jwt).

---

## Cost/Benefit Analysis

### Costs

| Cost | Estimate | Notes |
|------|----------|-------|
| **Labor** | 4-6 dev-days | 53 model files, 29 query sites, 10 API files |
| **Regression risk** | Medium | All DB queries rewrite; needs full integration test pass |
| **Tooling change** | `sqlc generate` added to build; CI pipeline update |
| **Learning curve** | Low | Team already writes raw SQL; sqlc is thin wrapper |
| **Migration risk** | Low | No schema changes; same DB; GORM already used as SQL passthrough |
| **Auto-migration loss** | Low-Medium | 53 `AutoMigrate` calls replaced by explicit DDL; needs migration tooling |
| **gorm.Model loss** | Low | ID/CreatedAt/UpdatedAt/DeletedAt become explicit struct fields |
| **Null type churn** | Medium | Custom `NullString`/`NullInt64`/`NullTime` replaced by `sql.Null*` or pointer types |
| **Picklist system** | Low | 25 picklist queries move from GORM named-param to sqlc `:many`; same SQL |

### Benefits

| Benefit | Impact | Notes |
|---------|--------|-------|
| **Compile-time SQL validation** | High | All queries validated against schema at build time — catches column renames, type mismatches, missing tables |
| **Type safety** | High | Generated structs match actual column types; no `Scan(&interface{})` guesswork |
| **Performance** | Low-Medium | Saves GORM reflection overhead; already mostly raw SQL though |
| **Smaller binary** | Low | Removes `gorm.io/gorm` + `gorm.io/driver/mysql` + `jinzhu/inflection` + `jinzhu/now` (~2MB) |
| **Fewer dependencies** | Medium | Drops 4 transitive deps; less supply-chain attack surface |
| **Code clarity** | High | Explicit SQL in `.sql` files, not string-concatenated in Go handlers |
| **Easier code review** | Medium | SQL changes are plain `.sql` diffs; Go changes are simple function calls |
| **Testing** | Medium | sqlc generates interfaces; easy to mock `Queries` for unit tests |
| **Standardization** | Medium | Matches sqlc+pgx stack used in other jbuchbinder projects (Concordia, A Penny Worth, Recevos) |

### Verdict: **Net Positive**

The codebase already treats GORM as a SQL passthrough — 79% of calls are `Raw()`. The
migration is closer to "add sqlc queries for existing SQL" than a full ORM rewrite.
Standardization across projects alone justifies the effort. Primary risk is regression
from the 6 ORM-style calls and the auto-migration → explicit migration shift, both
well-scoped.

---

## Phase 1: Infrastructure Setup

### Task 1.1: Install sqlc and create configuration

**Objective:** Set up sqlc tooling and project config.

**Files:**
- Create: `internal/db/sqlc.yaml`
- Create: `internal/db/schema.sql`
- Modify: `Makefile`

**Step 1: Install sqlc**

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

**Step 2: Create sqlc.yaml**

```yaml
# internal/db/sqlc.yaml
version: "2"
sql:
  - engine: "mysql"
    queries: "queries/"
    schema: "schema.sql"
    gen:
      go:
        package: "dbgen"
        out: "../../dbgen"
        emit_json_tags: true
        emit_pointers_for_null_types: true
```

**Step 3: Extract schema.sql from GORM models**

Write a script or manually compile the DDL for all 53 tables from the existing
FreeMED 0.9.x database into `internal/db/schema.sql`. Since the database already
exists and GORM auto-migrates, run:

```bash
mysqldump --no-data --skip-triggers --skip-add-drop-table freemed > internal/db/schema.sql
```

**Step 4: Add Makefile targets**

```makefile
sqlc:
	sqlc generate -f internal/db/sqlc.yaml
.PHONY: sqlc
```

**Verification:** `sqlc generate -f internal/db/sqlc.yaml` produces Go files in `dbgen/`.

---

### Task 1.2: Replace GORM connection with database/sql

**Objective:** Create a new connection manager using `database/sql` + `go-sql-driver/mysql`.

**Files:**
- Create: `internal/db/connection.go`
- Modify: `model/init.go` (replace `Db *gorm.DB` with `Db *sql.DB` or new interface)
- Modify: `cmd/freemed-server/main.go`

**Step 1: Write connection manager**

```go
// internal/db/connection.go
package db

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/freemed/freemed-server/config"
)

var Pool *sql.DB

func Open() (*sql.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true",
        config.Config.Database.User,
        config.Config.Database.Pass,
        config.Config.Database.Host,
        config.Config.Database.Name,
    )
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("db.Open: %w", err)
    }
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("db.Ping: %w", err)
    }
    Pool = db
    return db, nil
}
```

**Step 2: Transition model.Db**

Change `model/init.go`:
```go
var Db interface{} // Temporary interface{} during migration; final: *sql.DB
```

In `main.go`, replace `model.Db = model.InitDb()` with:
```go
sqlDB, err := db.Open()
if err != nil {
    panic(err)
}
model.Db = sqlDB
```

**Verification:** `go build ./...` compiles; server starts and connects to MySQL.

---

## Phase 2: Model Refactoring

### Task 2.1: Create sqlc query files for existing raw SQL

**Objective:** Extract raw SQL from API handlers into `.sql` files.

**Files:**
- Create: `internal/db/queries/patient.sql`
- Create: `internal/db/queries/scheduler.sql`
- Create: `internal/db/queries/support.sql`
- Create: `internal/db/queries/user.sql`

**Approach:** For each `model.Db.Raw(query, args...).Scan(&out)` call in API handlers:
1. Copy the SQL into a named query with `:many` or `:one` annotation
2. Map `?` placeholders to `sqlc.arg(name)` named parameters
3. Add `-- name: QueryName :many` header
4. Run `sqlc generate` to produce type-safe Go functions

**Example (patient.sql):**

```sql
-- name: PatientEmrAttachments :many
SELECT p.patient, p.module, p.oid, p.annotation, p.summary, p.stamp,
       DATE_FORMAT(p.stamp, '%m/%d/%Y') AS date_mdy,
       m.module_name AS module_name, m.module_class AS module_namespace,
       p.locked, p.id
FROM patient_emr p
LEFT OUTER JOIN modules m ON m.module_table = p.module
WHERE p.patient = sqlc.arg(patient_id) AND m.module_hidden = 0;

-- name: PatientEmrAttachmentsByModule :many
SELECT p.patient, p.module, p.oid, p.annotation, p.summary, p.stamp,
       DATE_FORMAT(p.stamp, '%m/%d/%Y') AS date_mdy,
       m.module_name AS module_name, m.module_class AS module_namespace,
       p.locked, p.id
FROM patient_emr p
LEFT OUTER JOIN modules m ON m.module_table = p.module
WHERE p.patient = sqlc.arg(patient_id)
  AND p.module = sqlc.arg(module)
  AND m.module_hidden = 0;
```

**Verification:** `sqlc generate` succeeds; generated Go types match existing struct fields.

---

### Task 2.2: Convert ORM-style queries to sqlc

**Objective:** Replace 6 GORM Find/First/Save calls with sqlc-generated functions.

**Files:**
- Modify: `model/user.go` (2 First, 1 Find in data_store)
- Modify: `api/data_store.go` (1 Find)
- Modify: `api/scheduler.go` (1 Save)
- Create: `internal/db/queries/data_store.sql`
- Create: `internal/db/queries/scheduler_crud.sql`

**Step 1: Write sqlc queries for each ORM call**

```sql
-- user.sql
-- name: GetUserByUsername :one
SELECT * FROM user WHERE username = sqlc.arg(username);

-- name: GetUserById :one
SELECT * FROM user WHERE id = sqlc.arg(id);

-- name: CheckUserPassword :one
SELECT * FROM user
WHERE username = sqlc.arg(username)
  AND userpassword = sqlc.arg(password);

-- data_store.sql
-- name: GetDataStoreByPatientModule :one
SELECT * FROM patient_data_store
WHERE patient = sqlc.arg(patient_id)
  AND module = LOWER(sqlc.arg(module))
  AND id = sqlc.arg(id);

-- scheduler_crud.sql
-- name: GetSchedulerById :one
SELECT * FROM scheduler WHERE id = sqlc.arg(id);

-- name: UpdateScheduler :exec
UPDATE scheduler SET
  caldateof = sqlc.arg(date_of),
  calhour = sqlc.arg(hour),
  calminute = sqlc.arg(minute),
  calduration = sqlc.arg(duration),
  calmodified = sqlc.arg(modified)
WHERE id = sqlc.arg(id);
```

**Step 2: Rewrite model/user.go**

Replace `GetUserByName`, `GetUserById`, `CheckUserPassword` to use sqlc-generated
`dbgen.New(model.Db.(*sql.DB))` or a passed-in `*dbgen.Queries`.

**Step 3: Rewrite api/scheduler.go reschedule handler**

Replace `model.Db.Save(&eventObj)` with `queries.UpdateScheduler(ctx, params)`.

**Verification:** `go build ./...` compiles; manual test of login, data_store GET, scheduler reschedule.

---

### Task 2.3: Remove gorm.Model from all 53 structs

**Objective:** Replace `gorm.Model` with explicit ID, CreatedAt, UpdatedAt, DeletedAt fields.

**Files:**
- Modify: all 53 model/*.go files

**Pattern:** Replace:
```go
type FooModel struct {
    gorm.Model
    Name string `db:"name" json:"name"`
}
```

With:
```go
type FooModel struct {
    ID        int64          `db:"id" json:"id"`
    CreatedAt time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
    DeletedAt gorm.DeletedAt `db:"deleted_at" json:"deleted_at"` // or sql.NullTime
    Name      string         `db:"name" json:"name"`
}
```

Note: The `db` struct tags are kept for compatibility with any remaining direct scanning.
sqlc-generated structs will provide their own types — the model structs become documentation
rather than active ORM entities.

**Verification:** `go build ./...` compiles after all 53 files updated and GORM import removed.

---

## Phase 3: API Handler Conversion

### Task 3.1: Refactor handlers to use sqlc Queries

**Objective:** Pass `*dbgen.Queries` (or `*sql.DB`) to handlers; replace `model.Db.Raw()` calls.

**Files:**
- Modify: `cmd/freemed-server/main.go`
- Modify: `api/patient.go`
- Modify: `api/scheduler.go`
- Modify: `api/support.go`
- Modify: `api/messages.go`
- Modify: `api/data_store.go`
- Modify: `api/remitt.go`
- Modify: `cmd/freemed-server/auth.go`

**Step 1: Add queries to handler init**

In `main.go`, create queries and pass to API registration:
```go
import "github.com/freemed/freemed-server/dbgen"

queries := dbgen.New(sqlDB)
```

Thread `queries` through to handlers. Since the current API uses `init()` functions
that register into a global `ApiMap`, add a `Queries` field to `ApiMapping`:

```go
type ApiMapping struct {
    Authenticated  bool
    Queries        *dbgen.Queries  // new field
    RouterFunction func(*gin.RouterGroup)
}
```

**Step 2: Replace each Raw() call**

Example (patientEmrAttachments):
```go
// Before:
tx := model.Db.Raw(query, id).Scan(&o)
err = tx.Error

// After:
rows, err := queries.PatientEmrAttachments(r.Context(), dbgen.PatientEmrAttachmentsParams{
    PatientID: patientID,
})
```

**Step 3: Handle picklist queries**

25 picklist queries stored as strings in `DbSupportPicklists` need conversion.
Since they're dynamically selected by name, either:
- Map them to sqlc-generated functions with a switch statement
- Or keep them as raw SQL on `*sql.DB` (acceptable for variable-shape picklists)

**Verification:** `go build ./...` compiles; manual test of each endpoint.

---

### Task 3.2: Remove GORM dependency entirely

**Objective:** Delete GORM imports, `InitDb()`, `AutoMigrate`, table registration, `gorm.Model`.

**Files:**
- Modify: `model/db.go` (remove or replace with `*sql.DB` passthrough)
- Modify: `model/init.go` (remove GORM import)
- Modify: all 53 model files (remove `gorm.io/gorm` import)
- Modify: `cmd/freemed-server/go.mod` (drop gorm.io/gorm, gorm.io/driver/mysql, jinzhu/*)

**Step 1: Remove `InitDb()`**

Delete `model/db.go` or replace with db-open delegation.

**Step 2: Remove `init()` table registrations**

Delete `DbTables = append(...)` from all 53 files. This also removes the
`DbSupportPicklists = append(...)` calls — those queries move to sqlc.

**Step 3: Remove `gorm.io/gorm` import**

From all model files and go.mod:
```bash
cd cmd/freemed-server
go mod edit -droprequire gorm.io/gorm
go mod edit -droprequire gorm.io/driver/mysql
# jinzhu/* are transitive and will drop automatically
go mod tidy
```

**Verification:** `go build ./...` compiles successfully with no GORM imports anywhere.

---

## Phase 4: Migration Tooling

### Task 4.1: Add database migration support

**Objective:** Replace GORM `AutoMigrate` with explicit versioned migrations.

**Files:**
- Create: `internal/db/migrations/000001_initial_schema.up.sql`
- Create: `internal/db/migrations/000001_initial_schema.down.sql`
- Modify: `Makefile`
- Modify: `cmd/freemed-server/main.go`

**Step 1: Generate initial migration from existing DB schema**

```bash
mysqldump --no-data --skip-triggers --skip-add-drop-table freemed \
  > internal/db/migrations/000001_initial_schema.up.sql
```

**Step 2: Install golang-migrate and add Makefile targets**

```makefile
migrate-up:
	migrate -path internal/db/migrations \
	  -database "mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST})/${DB_NAME}" up
migrate-down:
	migrate -path internal/db/migrations \
	  -database "mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST})/${DB_NAME}" down 1
```

**Step 3: Add startup migration check**

In `main.go`, if `config.Config.Database.Migrations == true`, run pending migrations
on startup (via `migrate` library import or subprocess).

**Verification:** `make migrate-up` applies schema to test database without errors.

---

## Phase 5: Testing & Cleanup

### Task 5.1: Write integration tests

**Objective:** Ensure all converted queries produce identical results.

**Files:**
- Create: `api/patient_test.go`
- Create: `api/scheduler_test.go`
- Create: `model/user_test.go`

**Approach:** For each converted query:
1. Set up test database with known data
2. Call old GORM code (if still present) and new sqlc code
3. Assert identical results
4. Remove old code once tests pass

**Verification:** `go test ./...` passes.

---

### Task 5.2: Final cleanup

**Objective:** Remove remaining GORM artifacts, update documentation.

**Files:**
- Modify: `README.md` (update tech stack section)
- Modify: `TODO.md` (remove GORM-related items if any)
- Delete: `model/db.go` (if not already removed)
- Modify: `config/config.go` (remove GORM-specific config if any)

**Verification:** `grep -r "gorm" --include="*.go" .` returns empty.

---

## Migration Order Summary

```
Phase 1: Infrastructure (sqlc tooling + connection manager)       [0.5 day]
Phase 2: Model refactoring (sql queries + gorm.Model removal)     [2.0 days]
  ├── 2.1: sqlc query files for 23 raw SQL calls                   [1.0 day]
  ├── 2.2: sqlc queries for 6 ORM calls                            [0.5 day]
  └── 2.3: Remove gorm.Model from 53 structs                       [0.5 day]
Phase 3: API handler conversion                                   [1.0 day]
  ├── 3.1: Wire sqlc Queries into 10 API files                     [0.5 day]
  └── 3.2: Remove GORM dependency entirely                         [0.5 day]
Phase 4: Migration tooling                                         [0.5 day]
Phase 5: Testing & cleanup                                         [1.0 day]
                                                      Total:     ~5.0 days
```

## Risk Mitigation

1. **Parallel run:** Keep GORM in a branch; compare sqlc query results against GORM results
   in staging before merging.
2. **Incremental merge:** Each phase produces a working server; merge phase-by-phase.
3. **Picklist regression:** The 25 picklist queries are the most diverse SQL shapes;
   test each one manually or write a picklist-verification script.
4. **MySQL `ONLY_FULL_GROUP_BY`:** Already documented as disabled; sqlc won't change this.
5. **Null type compatibility:** Custom `NullString`/`NullInt64`/`NullTime` are used in
   30+ files; sqlc with `emit_pointers_for_null_types: true` generates `*string`/`*int64`
   instead. Either keep custom types (with conversion helpers) or migrate to `*T` pointers.
