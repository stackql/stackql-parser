# Session configuration reset

SET and UNSET use configuration names on the left-hand side, without a dollar prefix.
The dollar sigil is reserved for dereferencing values. UNSET accepts one or more SQL identifiers:

```sql
UNSET "providers.production";
UNSET "auth.example.password", "secrets.token";
```

Identifiers use StackQL's double-quote syntax. UNSET is non-reserved, so a column can still be named
`unset`. Comments after UNSET are accepted. Values or assignments after a reset path are rejected.

The parser returns the existing `*sqlparser.Set` AST. Each path becomes a SetExpr whose value is
`*sqlparser.Default`. This is the same representation as `SET "providers.production" = DEFAULT`.
Preview classifies both spellings as StmtSet. The host owns configuration validation, authorization,
reset semantics and side effects; parsing executes nothing.

Provider bindings and secret configurations use existing SET string literals containing JSON.
Their object structure belongs to the host and Omni, not the SQL grammar. This keeps credentials
out of new AST fields and lets hosts reject unsupported configuration before applying it.

Run the focused regression with:

```sh
go test ./go/vt/sqlparser -run '^TestUnsetSessionConfiguration$'
```

After release, StackQL must consume the parser's new tag to accept scoped UNSET. Its existing tagged
parser already accepts unqualified UNSET, SET SESSION and SET ... = DEFAULT. No local module replacement is part of this feature.

## Optional scope

```sql
SET SESSION "auth.example.password" = 'value';
UNSET SESSION "auth.example.password", "providers.production";
```

Omitting scope leaves selection to the host (StackQL uses SESSION). SET and UNSET accept SESSION
and GLOBAL and record the chosen scope in each SetExpr. The host decides which scopes can execute;
StackQL currently rejects GLOBAL rather than treating it as a session mutation. Principal/group
storage, inheritance and conflict resolution are not implemented by this grammar.
