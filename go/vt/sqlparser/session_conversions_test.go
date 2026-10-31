package sqlparser

import "testing"

func TestSessionVariableConversionsAndIdentifiers(t *testing.T) {
	for _, statement := range []string{
		"select cast($some_id as text)",
		"select convert($some_id, text)",
		"select convert($digits, integer)",
		"select cast($some_id as bigint)",
		"select cast($enabled as boolean)",
		"select md5(lower(convert($some_id, text) || '-east'))",
		`SET "$.providers.prod" = '{"provider":"aws"}'`,
		`SET "providers.prod" = 'saved'`,
		`UNSET "$.providers.prod", ordinary_variable`,
		`select identifier($column) as "$label" from $provider.$service.$resource limit $count`,
	} {
		if _, err := Parse(statement); err != nil {
			t.Errorf("%s: %v", statement, err)
		}
	}
}

func TestConcatenationIsDistinctFromBooleanOr(t *testing.T) {
	statement, err := Parse("select id from widgets where region = convert($some_id, text) || '-east' or region = 'west'")
	if err != nil {
		t.Fatal(err)
	}
	selected := statement.(*Select)
	logical, ok := selected.Where.Expr.(*OrExpr)
	if !ok {
		t.Fatalf("expected OR predicate, got %T", selected.Where.Expr)
	}
	comparison, ok := logical.Left.(*ComparisonExpr)
	if !ok {
		t.Fatalf("expected comparison, got %T", logical.Left)
	}
	concat, ok := comparison.Right.(*BinaryExpr)
	if !ok || concat.Operator != "||" {
		t.Fatalf("expected concatenation on comparison RHS, got %T", comparison.Right)
	}
}
