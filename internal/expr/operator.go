package expr

import (
	"fmt"

	"github.com/chaisql/chai/internal/environment"
	"github.com/chaisql/chai/internal/object"
	"github.com/chaisql/chai/internal/sql/scanner"
	"github.com/chaisql/chai/internal/types"
	"github.com/cockroachdb/errors"
)

type simpleOperator struct {
	a, b Expr
	Tok  scanner.Token
}

func (op simpleOperator) Precedence() int {
	return op.Tok.Precedence()
}

func (op simpleOperator) LeftHand() Expr {
	return op.a
}

func (op simpleOperator) RightHand() Expr {
	return op.b
}

func (op *simpleOperator) SetLeftHandExpr(a Expr) {
	op.a = a
}

func (op *simpleOperator) SetRightHandExpr(b Expr) {
	op.b = b
}

func (op *simpleOperator) Token() scanner.Token {
	return op.Tok
}

func (op *simpleOperator) eval(env *environment.Environment, fn func(a, b types.Value) (types.Value, error)) (types.Value, error) {
	if op.a == nil || op.b == nil {
		return NullLiteral, errors.New("missing operand")
	}

	va, err := op.a.Eval(env)
	if err != nil {
		return NullLiteral, err
	}

	vb, err := op.b.Eval(env)
	if err != nil {
		return NullLiteral, err
	}

	return fn(va, vb)
}

// Equal compares this expression with the other expression and returns
// true if they are equal.
func (op *simpleOperator) IsEqual(other Expr) bool {
	if other == nil {
		return false
	}

	oop, ok := other.(Operator)
	if !ok {
		return false
	}

	return op.Tok == oop.Token() &&
		Equal(op.a, oop.LeftHand()) &&
		Equal(op.b, oop.RightHand())
}

func (op *simpleOperator) String() string {
	return fmt.Sprintf("%v %v %v", op.a, op.Tok, op.b)
}

// An Operator is a binary expression that
// takes two operands and executes an operation on them.
type Operator interface {
	Expr

	Precedence() int
	LeftHand() Expr
	RightHand() Expr
	SetLeftHandExpr(Expr)
	SetRightHandExpr(Expr)
	Token() scanner.Token
}

type ConcatOperator struct {
	*simpleOperator
}

// Concat creates an expression that concatenates two text values together.
// It returns null if one of the values is not a text.
func Concat(a, b Expr) Expr {
	return &ConcatOperator{&simpleOperator{a, b, scanner.CONCAT}}
}

func (op *ConcatOperator) Eval(env *environment.Environment) (types.Value, error) {
	return op.simpleOperator.eval(env, func(a, b types.Value) (types.Value, error) {
		if a.Type() != types.TypeText || b.Type() != types.TypeText {
			return NullLiteral, nil
		}

		return types.NewTextValue(types.AsString(a) + types.AsString(b)), nil
	})
}

// Cast represents the CAST expression.
type Cast struct {
	Expr   Expr
	CastAs types.Type
}

// Eval returns the primary key of the current object.
func (c Cast) Eval(env *environment.Environment) (types.Value, error) {
	v, err := c.Expr.Eval(env)
	if err != nil {
		return v, err
	}

	return object.CastAs(v, c.CastAs)
}

// IsEqual compares this expression with the other expression and returns
// true if they are equal.
func (c Cast) IsEqual(other Expr) bool {
	if other == nil {
		return false
	}

	o, ok := other.(Cast)
	if !ok {
		return false
	}

	if c.CastAs != o.CastAs {
		return false
	}

	if c.Expr != nil {
		return Equal(c.Expr, o.Expr)
	}

	return o.Expr != nil
}

func (c Cast) Params() []Expr { return []Expr{c.Expr} }

func (c Cast) String() string {
	return fmt.Sprintf("CAST(%v AS %v)", c.Expr, c.CastAs)
}
