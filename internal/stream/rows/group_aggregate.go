package rows

import (
	"fmt"
	"strings"

	"github.com/chaisql/chai/internal/environment"
	"github.com/chaisql/chai/internal/expr"
	"github.com/chaisql/chai/internal/object"
	"github.com/chaisql/chai/internal/stream"
	"github.com/chaisql/chai/internal/types"
)

type GroupAggregateOperator struct {
	stream.BaseOperator
	Builders []expr.AggregatorBuilder
	E        expr.Expr
}

// GroupAggregate consumes the incoming stream and outputs one value per group.
// It assumes the stream is sorted by groupBy.
func GroupAggregate(groupBy expr.Expr, builders ...expr.AggregatorBuilder) *GroupAggregateOperator {
	return &GroupAggregateOperator{E: groupBy, Builders: builders}
}

func (op *GroupAggregateOperator) Iterate(in *environment.Environment, f func(out *environment.Environment) error) error {
	var lastGroup types.Value
	var ga *groupAggregator

	var groupExpr string
	if op.E != nil {
		groupExpr = op.E.String()
	}

	err := op.Prev.Iterate(in, func(out *environment.Environment) error {
		if op.E == nil {
			if ga == nil {
				ga = newGroupAggregator(nil, groupExpr, op.Builders)
			}

			return ga.Aggregate(out)
		}

		group, err := op.E.Eval(out)
		if err != nil {
			return err
		}

		// handle the first object of the stream
		if lastGroup == nil {
			lastGroup, err = object.CloneValue(group)
			if err != nil {
				return err
			}
			ga = newGroupAggregator(lastGroup, groupExpr, op.Builders)
			return ga.Aggregate(out)
		}

		ok, err := lastGroup.EQ(group)
		if err != nil {
			return err
		}
		if ok {
			return ga.Aggregate(out)
		}

		// if the object is from a different group, we flush the previous group, emit it and start a new group
		e, err := ga.Flush(out)
		if err != nil {
			return err
		}
		err = f(e)
		if err != nil {
			return err
		}

		lastGroup, err = object.CloneValue(group)
		if err != nil {
			return err
		}

		ga = newGroupAggregator(lastGroup, groupExpr, op.Builders)
		return ga.Aggregate(out)
	})
	if err != nil {
		return err
	}

	// if s is empty, we create a default group so that aggregators will
	// return their default initial value.
	// Ex: For `SELECT COUNT(*) FROM foo`, if `foo` is empty
	// we want the following result:
	// {"COUNT(*)": 0}
	if ga == nil {
		ga = newGroupAggregator(nil, "", op.Builders)
	}

	e, err := ga.Flush(in)
	if err != nil {
		return err
	}
	return f(e)
}

func (op *GroupAggregateOperator) String() string {
	var sb strings.Builder

	sb.WriteString("rows.GroupAggregate(")
	if op.E != nil {
		sb.WriteString(op.E.String())
	} else {
		sb.WriteString("NULL")
	}

	for _, agg := range op.Builders {
		sb.WriteString(", ")
		sb.WriteString(agg.(fmt.Stringer).String())
	}

	sb.WriteString(")")
	return sb.String()
}

// a groupAggregator is an aggregator for a whole group of objects.
// It applies all the aggregators for each objects and returns a new object with the
// result of the aggregation.
type groupAggregator struct {
	group       types.Value
	groupExpr   string
	aggregators []expr.Aggregator
}

func newGroupAggregator(group types.Value, groupExpr string, builders []expr.AggregatorBuilder) *groupAggregator {
	newAggregators := make([]expr.Aggregator, len(builders))
	for i, b := range builders {
		newAggregators[i] = b.Aggregator()
	}

	return &groupAggregator{
		aggregators: newAggregators,
		group:       group,
		groupExpr:   groupExpr,
	}
}

func (g *groupAggregator) Aggregate(env *environment.Environment) error {
	for _, agg := range g.aggregators {
		err := agg.Aggregate(env)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *groupAggregator) Flush(env *environment.Environment) (*environment.Environment, error) {
	fb := object.NewFieldBuffer()

	// add the current group to the object
	if g.groupExpr != "" {
		fb.Add(g.groupExpr, g.group)
	}

	for _, agg := range g.aggregators {
		v, err := agg.Eval(env)
		if err != nil {
			return nil, err
		}
		fb.Add(agg.String(), v)
	}

	var newEnv environment.Environment
	newEnv.SetOuter(env)
	newEnv.SetRowFromObject(fb)

	return &newEnv, nil
}
