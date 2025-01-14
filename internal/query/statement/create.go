package statement

import (
	"math"

	"github.com/chaisql/chai/internal/database"
	errs "github.com/chaisql/chai/internal/errors"
	"github.com/chaisql/chai/internal/stream"
	"github.com/chaisql/chai/internal/stream/index"
	"github.com/chaisql/chai/internal/stream/table"
)

// CreateTableStmt represents a parsed CREATE TABLE statement.
type CreateTableStmt struct {
	IfNotExists bool
	Info        database.TableInfo
}

// IsReadOnly always returns false. It implements the Statement interface.
func (stmt *CreateTableStmt) IsReadOnly() bool {
	return false
}

// Run runs the Create table statement in the given transaction.
// It implements the Statement interface.
func (stmt *CreateTableStmt) Run(ctx *Context) (Result, error) {
	var res Result

	// if there is no primary key, create a rowid sequence
	if stmt.Info.PrimaryKey == nil {
		seq := database.SequenceInfo{
			IncrementBy: 1,
			Min:         1, Max: math.MaxInt64,
			Start: 1,
			Cache: 64,
			Owner: database.Owner{
				TableName: stmt.Info.TableName,
			},
		}
		err := ctx.Tx.CatalogWriter().CreateSequence(ctx.Tx, &seq)
		if err != nil {
			return res, err
		}

		stmt.Info.RowidSequenceName = seq.Name
	}

	err := ctx.Tx.CatalogWriter().CreateTable(ctx.Tx, stmt.Info.TableName, &stmt.Info)
	if stmt.IfNotExists {
		if errs.IsAlreadyExistsError(err) {
			return res, nil
		}
	}

	// create a unique index for every unique constraint
	for _, tc := range stmt.Info.TableConstraints {
		if tc.Unique {
			_, err = ctx.Tx.CatalogWriter().CreateIndex(ctx.Tx, &database.IndexInfo{
				Paths:  tc.Paths,
				Unique: true,
				Owner: database.Owner{
					TableName: stmt.Info.TableName,
					Paths:     tc.Paths,
				},
				KeySortOrder: tc.SortOrder,
			})
			if err != nil {
				return res, err
			}
		}
	}

	return res, err
}

// CreateIndexStmt represents a parsed CREATE INDEX statement.
type CreateIndexStmt struct {
	IfNotExists bool
	Info        database.IndexInfo
}

// IsReadOnly always returns false. It implements the Statement interface.
func (stmt *CreateIndexStmt) IsReadOnly() bool {
	return false
}

// Run runs the Create index statement in the given transaction.
// It implements the Statement interface.
func (stmt *CreateIndexStmt) Run(ctx *Context) (Result, error) {
	var res Result

	_, err := ctx.Tx.CatalogWriter().CreateIndex(ctx.Tx, &stmt.Info)
	if stmt.IfNotExists {
		if errs.IsAlreadyExistsError(err) {
			return res, nil
		}
	}
	if err != nil {
		return res, err
	}

	s := stream.New(table.Scan(stmt.Info.Owner.TableName)).
		Pipe(index.Insert(stmt.Info.IndexName)).
		Pipe(stream.Discard())

	ss := PreparedStreamStmt{
		Stream:   s,
		ReadOnly: false,
	}

	return ss.Run(ctx)
}

// CreateSequenceStmt represents a parsed CREATE SEQUENCE statement.
type CreateSequenceStmt struct {
	IfNotExists bool
	Info        database.SequenceInfo
}

// IsReadOnly always returns false. It implements the Statement interface.
func (stmt *CreateSequenceStmt) IsReadOnly() bool {
	return false
}

// Run the statement in the given transaction.
// It implements the Statement interface.
func (stmt *CreateSequenceStmt) Run(ctx *Context) (Result, error) {
	var res Result

	err := ctx.Tx.CatalogWriter().CreateSequence(ctx.Tx, &stmt.Info)
	if stmt.IfNotExists {
		if errs.IsAlreadyExistsError(err) {
			return res, nil
		}
	}
	return res, err
}
