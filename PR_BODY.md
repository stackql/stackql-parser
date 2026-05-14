## Summary

Demotes Vitess-inherited reserved keywords that block stackql users from using common identifiers (`force`, `match`, `rank`, `system`, `unique`, etc.) as column/parameter names. The trigger was `DELETE ... WHERE force = true` failing on Databricks queries.

Three phases, each its own commit:

- **Phase 1** - 20 bucket-C tokens (declared reserved but never emitted by the lexer; purely cosmetic):
  `ARRAY  CUME_DIST  DENSE_RANK  FIRST_VALUE  GROUPING  GROUPS  JSON_TABLE  LAG  LAST_VALUE  LATERAL  LEAD  MEMBER  NTH_VALUE  NTILE  OF  PERCENT_RANK  RANK  ROW_NUMBER  SYSTEM  WINDOW`
- **Phase 2** - 7 MySQL-only mid-clause tokens:
  `FORCE  MATCH  MAXVALUE  DISTINCTROW  STRAIGHT_JOIN  UNLOCK  AUTO_INCREMENT`
- **Phase 3** - 2 demoted + 1 reverted:
  Demoted: `RENAME`, `UNIQUE`. Reverted: `LOCK` (see Behavior changes).

## y.output conflict count delta

| stage | shift/reduce | reduce/reduce |
|---|---|---|
| baseline (main) | 209 | 470 |
| after Phase 1 | 229 (+20) | 510 (+40) |
| after Phase 2 | 487 (+258) | 523 (+13) |
| after Phase 3 | 491 (+4) | 528 (+5) |

Phase 1 conflicts are all of the form `conflict on <TOKEN>` for the 20 tokens - the trivial new reduction path from `non_reserved_keyword` to `sql_id`. None of those tokens are in the lexer `keywords` map in `token.go`, so they cannot be reached at runtime.

Phase 2's larger jump is mostly on `DISTINCTROW` and `STRAIGHT_JOIN`, plus one new shift/reduce on `NOT` from the MATCH-after-value-expression context. goyacc's default shift-over-reduce resolution keeps every previously-exercised productive rule working except `FORCE INDEX` (see below).

Phase 3 is tight: only RENAME and UNIQUE were kept, adding 4+5 conflicts on those tokens.

## Behavior changes

### Gained (intended)

The 29 demoted keywords can now be used as unquoted column and parameter identifiers. The original bug repro now parses:

```sql
DELETE FROM databricks_workspace.catalog.storage_credentials
WHERE name = '...' AND deployment_name = '...' AND force = true;
```

### Lost (documented)

- `SELECT * FROM t FORCE INDEX (idx)` no longer parses. stackql does not implement indexes; this Vitess hint production is dead code in stackql. Users who relied on the parser accepting MySQL `FORCE INDEX` hints (rare in stackql usage) will need to remove them.

### Phase 3 revert

`LOCK` was attempted but reverted because demoting it broke `SELECT * FROM t LOCK IN SHARE MODE` (state 1022 shift/reduce on LOCK resolves toward `sql_id`, not the lock-mode suffix production). stackql does not implement row-level locking, but the phase's strict conflict-verification rule says revert; honoring that.

## Tests

New `TestDemotedReservedKeywords` in `go/vt/sqlparser/parse_test.go` with 13 fixture rows grouped by phase. Each phase's fixtures pass after its commit.

Across the whole package: zero regressions among the 45 tests that were passing on `main`. The 26 pre-existing failures on `main` (`TestKeywords`, `TestValid`, `TestCaseSensitivity`, etc., all showing identifier output as `"dual"` instead of `dual` due to a fork-level rewrite difference) continue to fail with identical output - they are unrelated to this work.

## Files touched

- `go/vt/sqlparser/sql.y` - the only hand-edited grammar file (~58 lines moved).
- `go/vt/sqlparser/sql.go` - regenerated via `go run golang.org/x/tools/cmd/goyacc -o sql.go sql.y` per the existing Makefile.
- `go/vt/sqlparser/parse_test.go` - new test function.

`token.go` was deliberately NOT modified - the lexer keyword map is unchanged.

## Test plan

- [x] Each phase regenerates `sql.go` cleanly via the Makefile invocation.
- [x] `go build ./go/vt/sqlparser/...` succeeds after each phase.
- [x] `TestDemotedReservedKeywords` passes after each phase (filtered by `-run`).
- [x] No previously-passing test in the sqlparser package regresses (verified by diffing PASS sets vs `git stash`-ed baseline).
- [x] Productive-rule probes for every demoted token's existing grammar use (excluding the one documented FORCE INDEX regression).
