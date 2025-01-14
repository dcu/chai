package rows_test

import (
	"testing"

	"github.com/chaisql/chai/internal/environment"
	"github.com/chaisql/chai/internal/expr"
	"github.com/chaisql/chai/internal/sql/parser"
	"github.com/chaisql/chai/internal/stream"
	"github.com/chaisql/chai/internal/stream/rows"
	"github.com/chaisql/chai/internal/testutil"
	"github.com/chaisql/chai/internal/testutil/assert"
	"github.com/chaisql/chai/internal/types"
	"github.com/stretchr/testify/require"
)

func TestRowsEmit(t *testing.T) {
	tests := []struct {
		e      expr.Expr
		output types.Object
		fails  bool
	}{
		{parser.MustParseExpr("3 + 4"), nil, true},
		{parser.MustParseExpr("{a: 3 + 4}"), testutil.MakeObject(t, `{"a": 7}`), false},
	}

	for _, test := range tests {
		t.Run(test.e.String(), func(t *testing.T) {
			s := stream.New(rows.Emit(test.e))

			err := s.Iterate(new(environment.Environment), func(env *environment.Environment) error {
				r, ok := env.GetRow()
				require.True(t, ok)
				require.Equal(t, r.Object(), test.output)
				return nil
			})
			if test.fails {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("String", func(t *testing.T) {
		require.Equal(t, rows.Emit(parser.MustParseExpr("1 + 1"), parser.MustParseExpr("pk()")).String(), "rows.Emit(1 + 1, pk())")
	})
}
