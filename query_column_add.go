package bun

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/j2gg0s/bun/internal"
	"github.com/j2gg0s/bun/schema"
)

type AddColumnQuery struct {
	baseQuery
}

func NewAddColumnQuery(db *DB) *AddColumnQuery {
	q := &AddColumnQuery{
		baseQuery: baseQuery{
			db:   db,
			conn: db.DB,
		},
	}
	return q
}

func (q *AddColumnQuery) Conn(db IConn) *AddColumnQuery {
	q.setConn(db)
	return q
}

func (q *AddColumnQuery) Model(model interface{}) *AddColumnQuery {
	q.setTableModel(model)
	return q
}

//------------------------------------------------------------------------------

func (q *AddColumnQuery) Table(tables ...string) *AddColumnQuery {
	for _, table := range tables {
		q.addTable(schema.UnsafeIdent(table))
	}
	return q
}

func (q *AddColumnQuery) TableExpr(query string, args ...interface{}) *AddColumnQuery {
	q.addTable(schema.SafeQuery(query, args))
	return q
}

func (q *AddColumnQuery) ModelTableExpr(query string, args ...interface{}) *AddColumnQuery {
	q.modelTable = schema.SafeQuery(query, args)
	return q
}

//------------------------------------------------------------------------------

func (q *AddColumnQuery) ColumnExpr(query string, args ...interface{}) *AddColumnQuery {
	q.addColumn(schema.SafeQuery(query, args))
	return q
}

//------------------------------------------------------------------------------

func (q *AddColumnQuery) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}
	if len(q.columns) != 1 {
		return nil, fmt.Errorf("bun: AddColumnQuery requires exactly one column")
	}

	b = append(b, "ALTER TABLE "...)

	b, err = q.appendFirstTable(fmter, b)
	if err != nil {
		return nil, err
	}

	b = append(b, " ADD "...)

	b, err = q.columns[0].AppendQuery(fmter, b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

//------------------------------------------------------------------------------

func (q *AddColumnQuery) Exec(ctx context.Context, dest ...interface{}) (sql.Result, error) {
	queryBytes, err := q.AppendQuery(q.db.fmter, q.db.makeQueryBytes())
	if err != nil {
		return nil, err
	}

	query := internal.String(queryBytes)

	res, err := q.exec(ctx, q, query)
	if err != nil {
		return nil, err
	}

	return res, nil
}
