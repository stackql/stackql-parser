package sqlparser

import "testing"

func TestUnsetSessionConfiguration(t *testing.T) {
	for _, sql := range []string{
		"unset \"$.auth.aws.password\"",
		"unset password",
		"unset /* reset */ \"$.providers.production\", \"$.secrets.token\"",
	} {
		statement, err := Parse(sql)
		if err != nil {
			t.Fatal(err)
		}
		assignment, ok := statement.(*Set)
		if !ok {
			t.Fatalf("got %T", statement)
		}
		for _, expr := range assignment.Exprs {
			if _, ok := expr.Expr.(*Default); !ok {
				t.Fatalf("unset must normalize to DEFAULT: %T", expr.Expr)
			}
		}
		for _, expr := range assignment.Exprs {
			if expr.Name.GetRawVal() == "" {
				t.Fatal("missing reset path")
			}
		}
		if Preview(sql) != StmtSet {
			t.Fatal("unset must route as session configuration")
		}
	}
	for _, sql := range []string{"unset", "unset \"$.auth.aws\" = 'value'", "unset \"$.auth.aws\","} {
		if _, err := Parse(sql); err == nil {
			t.Fatalf("accepted %q", sql)
		}
	}
	if _, err := Parse("select unset from t"); err != nil {
		t.Fatal(err)
	}
}
