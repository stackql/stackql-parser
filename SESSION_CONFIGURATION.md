# Session configuration reset

UNSET accepts one or more SQL identifiers naming configuration paths:

```sql
UNSET "$.providers.production";
UNSET "$.auth.example.password", "$.secrets.token";
```

Identifiers use StackQL's double-quote syntax. UNSET is non-reserved, so a column can still be named
`unset`. Comments after UNSET are accepted. Values or assignments after a reset path are rejected.

The parser returns the existing `*sqlparser.Set` AST. Each path becomes a SetExpr whose value is
`*sqlparser.Default`. This is the same representation as `SET "$.providers.production" = DEFAULT`.
Preview classifies both spellings as StmtSet. The host owns configuration validation, authorization,
reset semantics and side effects; parsing executes nothing.

Provider bindings and secret configurations use existing SET string literals containing JSON.
Their object structure belongs to the host and Omni, not the SQL grammar. This keeps credentials
out of new AST fields and lets hosts reject unsupported configuration before applying it.

Run the focused regression with:

```sh
go test ./go/vt/sqlparser -run '^TestUnsetSessionConfiguration$'
```

After release, StackQL must consume the parser's new tag to accept UNSET. Its existing tagged parser
can already parse SET ... = DEFAULT. No local module replacement is part of this feature.
