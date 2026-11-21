package sqlparser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecBoolLiteral(t *testing.T) {
	for _, q := range []string{
		"EXEC m.t @flag = true",
		"EXEC m.t @flag = false",
		"EXEC m.t @flag = TRUE",
		"EXEC m.t @flag = FALSE",
	} {
		stmt, err := Parse(q)
		require.NoError(t, err, q)
		require.IsType(t, &Exec{}, stmt)
	}
}
EOF