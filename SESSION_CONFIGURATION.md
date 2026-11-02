# Session configuration syntax

SET and UNSET parse SQL identifiers, including a leading dollar sign. The parser records the
identifier and optional scope; StackQL owns the distinction between variables and runtime config.

```sql
SET region = 'ap-southeast-2';
SET "providers.saved" = '{"provider":"aws"}';
SET "$.auth.example.password" = 'session-password';
SET SESSION "$.providers.production" = '{"provider":"aws","tenancy":"production","auth":{"type":"aws_signing_v4"}}';
UNSET region, "providers.saved";
UNSET SESSION "$.auth.example.password", "$.providers.production";
```

In StackQL, a dollar-prefixed assignment/reset target selects intrinsic runtime configuration
for this session. A plain target selects a session scalar variable; dots alone do not select
configuration. A dollar-prefixed expression dereferences a value. These are host semantics, not
parser-managed storage. GLOBAL parses but the host currently rejects unsupported shared scopes.

UNSET becomes a Set statement whose expressions use Default, matching SET name = DEFAULT.
SET and UNSET retain the existing identifier/quoting grammar. Single-quoted '$name' is a string
literal, not a dereference. Object-name resolution is context-sensitive host AST processing.

CAST(expr AS TEXT/INTEGER/INT/BIGINT/BOOL/BOOLEAN) and corresponding CONVERT(expr, type) forms
are accepted in addition to the existing conversion types. Both forms produce ConvertExpr.
The || token produces BinaryExpr concatenation, distinct from OrExpr for boolean OR; it binds
more tightly than comparisons. Formatting preserves that distinction and expression grouping.

The consumer must release/pin the parser changes before relying on these conversion and
concatenation fixes. Local development replacements must not be committed as dependency pins.
