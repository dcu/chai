package dbutil

import (
	"fmt"
	"io"

	"github.com/chaisql/chai"
	"go.uber.org/multierr"
)

// Dump takes a database and dumps its content as SQL queries in the given writer.
// If tables is provided, only selected tables will be outputted.
func Dump(db *chai.DB, w io.Writer, tables ...string) error {
	tx, err := db.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = fmt.Fprintln(w, "BEGIN TRANSACTION;"); err != nil {
		return err
	}

	i := 0
	err = QueryTables(tx, tables, func(name, query string) error {
		// Blank separation between tables.
		if i > 0 {
			if _, err := fmt.Fprintln(w, ""); err != nil {
				return err
			}
		}
		i++

		return dumpTable(tx, w, query, name)
	})
	if err != nil {
		_, er := fmt.Fprintln(w, "ROLLBACK;")
		return multierr.Append(err, er)
	}

	_, err = fmt.Fprintln(w, "COMMIT;")
	return err
}

// dumpTable displays the content of the given table as SQL statements.
func dumpTable(tx *chai.Tx, w io.Writer, query, tableName string) error {
	// Dump schema first.
	if err := dumpSchema(tx, w, query, tableName); err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT * FROM %s", tableName)
	res, err := tx.Query(q)
	if err != nil {
		return err
	}
	defer res.Close()

	// Inserts statements.
	insert := fmt.Sprintf("INSERT INTO %s VALUES", tableName)
	return res.Iterate(func(r *chai.Row) error {
		data, err := r.MarshalJSON()
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "%s %s;\n", insert, string(data)); err != nil {
			return err
		}

		return nil
	})
}

// DumpSchema takes a database and dumps its schema as SQL queries in the given writer.
// If tables are provided, only selected tables will be outputted.
func DumpSchema(db *chai.DB, w io.Writer, tables ...string) error {
	tx, err := db.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	i := 0
	return QueryTables(tx, tables, func(name, query string) error {
		// Blank separation between tables.
		if i > 0 {
			if _, err := fmt.Fprintln(w, ""); err != nil {
				return err
			}
		}
		i++

		return dumpSchema(tx, w, query, name)
	})
}

// dumpSchema displays the schema of the given table as SQL statements.
func dumpSchema(tx *chai.Tx, w io.Writer, query string, tableName string) error {
	_, err := fmt.Fprintf(w, "%s;\n", query)
	if err != nil {
		return err
	}

	// Indexes statements.
	res, err := tx.Query(`
		SELECT sql FROM __chai_catalog WHERE 
			type = 'index' AND owner.table_name = ? OR
			type = 'sequence' AND owner IS NULL
	`, tableName)
	if err != nil {
		return err
	}
	defer res.Close()

	return res.Iterate(func(r *chai.Row) error {
		var q string

		err = r.Scan(&q)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(w, "%s;\n", q)
		return err
	})
}
