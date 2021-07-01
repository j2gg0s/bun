package bun

import (
	"context"
	"reflect"

	"github.com/j2gg0s/bun/internal"
	"github.com/j2gg0s/bun/schema"
)

type join struct {
	Parent    *join
	BaseModel tableModel
	JoinModel tableModel
	Relation  *schema.Relation

	ApplyQueryFunc func(*SelectQuery) *SelectQuery
	columns        []schema.QueryWithArgs
}

func (j *join) applyQuery(q *SelectQuery) {
	if j.ApplyQueryFunc == nil {
		return
	}

	var table *schema.Table
	var columns []schema.QueryWithArgs

	// Save state.
	table, q.table = q.table, j.JoinModel.Table()
	columns, q.columns = q.columns, nil

	q = j.ApplyQueryFunc(q)

	// Restore state.
	q.table = table
	j.columns, q.columns = q.columns, columns
}

func (j *join) Select(ctx context.Context, q *SelectQuery) error {
	switch j.Relation.Type {
	case schema.HasManyRelation:
		return j.selectMany(ctx, q)
	case schema.ManyToManyRelation:
		return j.selectM2M(ctx, q)
	}
	panic("not reached")
}

func (j *join) selectMany(ctx context.Context, q *SelectQuery) error {
	q = j.manyQuery(q)
	if q == nil {
		return nil
	}
	return q.Scan(ctx)
}

func (j *join) manyQuery(q *SelectQuery) *SelectQuery {
	hasManyModel := newHasManyModel(j)
	if hasManyModel == nil {
		return nil
	}

	q = q.Model(hasManyModel)

	var where []byte
	if len(j.Relation.JoinFields) > 1 {
		where = append(where, '(')
	}
	where = appendColumns(where, j.JoinModel.Table().SQLAlias, j.Relation.JoinFields)
	if len(j.Relation.JoinFields) > 1 {
		where = append(where, ')')
	}
	where = append(where, " IN ("...)
	where = appendChildValues(
		q.db.Formatter(),
		where,
		j.JoinModel.Root(),
		j.JoinModel.ParentIndex(),
		j.Relation.BaseFields,
	)
	where = append(where, ")"...)
	q = q.Where(internal.String(where))

	if j.Relation.PolymorphicField != nil {
		q = q.Where("? = ?", j.Relation.PolymorphicField.SQLName, j.Relation.PolymorphicValue)
	}

	j.applyQuery(q)
	q = q.Apply(j.hasManyColumns)

	return q
}

func (j *join) hasManyColumns(q *SelectQuery) *SelectQuery {
	if j.Relation.M2MTable != nil {
		q = q.ColumnExpr(string(j.Relation.M2MTable.SQLAlias) + ".*")
	}

	b := make([]byte, 0, 32)

	if len(j.columns) > 0 {
		for i, col := range j.columns {
			if i > 0 {
				b = append(b, ", "...)
			}

			var err error
			b, err = col.AppendQuery(q.db.fmter, b)
			if err != nil {
				q.err = err
				return q
			}
		}
	} else {
		joinTable := j.JoinModel.Table()
		b = appendColumns(b, joinTable.SQLAlias, joinTable.Fields)
	}

	q = q.ColumnExpr(internal.String(b))

	return q
}

func (j *join) selectM2M(ctx context.Context, q *SelectQuery) error {
	q = j.m2mQuery(q)
	if q == nil {
		return nil
	}
	return q.Scan(ctx)
}

func (j *join) m2mQuery(q *SelectQuery) *SelectQuery {
	fmter := q.db.fmter

	m2mModel := newM2MModel(j)
	if m2mModel == nil {
		return nil
	}
	q = q.Model(m2mModel)

	index := j.JoinModel.ParentIndex()
	baseTable := j.BaseModel.Table()

	//nolint
	var join []byte
	join = append(join, "JOIN "...)
	join = fmter.AppendQuery(join, string(j.Relation.M2MTable.Name))
	join = append(join, " AS "...)
	join = append(join, j.Relation.M2MTable.SQLAlias...)
	join = append(join, " ON ("...)
	for i, col := range j.Relation.M2MBaseFields {
		if i > 0 {
			join = append(join, ", "...)
		}
		join = append(join, j.Relation.M2MTable.SQLAlias...)
		join = append(join, '.')
		join = append(join, col.SQLName...)
	}
	join = append(join, ") IN ("...)
	join = appendChildValues(fmter, join, j.BaseModel.Root(), index, baseTable.PKs)
	join = append(join, ")"...)
	q = q.Join(internal.String(join))

	joinTable := j.JoinModel.Table()
	for i, m2mJoinField := range j.Relation.M2MJoinFields {
		joinField := j.Relation.JoinFields[i]
		q = q.Where("?.? = ?.?",
			joinTable.SQLAlias, joinField.SQLName,
			j.Relation.M2MTable.SQLAlias, m2mJoinField.SQLName)
	}

	j.applyQuery(q)
	q = q.Apply(j.hasManyColumns)

	return q
}

func (j *join) hasParent() bool {
	if j.Parent != nil {
		switch j.Parent.Relation.Type {
		case schema.HasOneRelation, schema.BelongsToRelation:
			return true
		}
	}
	return false
}

func (j *join) appendAlias(fmter schema.Formatter, b []byte) []byte {
	quote := fmter.IdentQuote()

	b = append(b, quote)
	b = appendAlias(b, j)
	b = append(b, quote)
	return b
}

func (j *join) appendAliasColumn(fmter schema.Formatter, b []byte, column string) []byte {
	quote := fmter.IdentQuote()

	b = append(b, quote)
	b = appendAlias(b, j)
	b = append(b, "__"...)
	b = append(b, column...)
	b = append(b, quote)
	return b
}

func (j *join) appendBaseAlias(fmter schema.Formatter, b []byte) []byte {
	quote := fmter.IdentQuote()

	if j.hasParent() {
		b = append(b, quote)
		b = appendAlias(b, j.Parent)
		b = append(b, quote)
		return b
	}
	return append(b, j.BaseModel.Table().SQLAlias...)
}

func (j *join) appendSoftDelete(b []byte, flags internal.Flag) []byte {
	b = append(b, '.')
	b = append(b, j.JoinModel.Table().SoftDeleteField.SQLName...)
	if flags.Has(deletedFlag) {
		b = append(b, " IS NOT NULL"...)
	} else {
		b = append(b, " IS NULL"...)
	}
	return b
}

func appendAlias(b []byte, j *join) []byte {
	if j.hasParent() {
		b = appendAlias(b, j.Parent)
		b = append(b, "__"...)
	}
	b = append(b, j.Relation.Field.Name...)
	return b
}

func (j *join) appendHasOneJoin(
	fmter schema.Formatter, b []byte, q *SelectQuery,
) (_ []byte, err error) {
	isSoftDelete := j.JoinModel.Table().SoftDeleteField != nil && !q.flags.Has(allWithDeletedFlag)

	b = append(b, "LEFT JOIN "...)
	b = fmter.AppendQuery(b, string(j.JoinModel.Table().SQLNameForSelects))
	b = append(b, " AS "...)
	b = j.appendAlias(fmter, b)

	b = append(b, " ON "...)

	b = append(b, '(')
	for i, baseField := range j.Relation.BaseFields {
		if i > 0 {
			b = append(b, " AND "...)
		}
		b = j.appendAlias(fmter, b)
		b = append(b, '.')
		b = append(b, j.Relation.JoinFields[i].SQLName...)
		b = append(b, " = "...)
		b = j.appendBaseAlias(fmter, b)
		b = append(b, '.')
		b = append(b, baseField.SQLName...)
	}
	b = append(b, ')')

	if isSoftDelete {
		b = append(b, " AND "...)
		b = j.appendAlias(fmter, b)
		b = j.appendSoftDelete(b, q.flags)
	}

	return b, nil
}

func appendChildValues(
	fmter schema.Formatter, b []byte, v reflect.Value, index []int, fields []*schema.Field,
) []byte {
	seen := make(map[string]struct{})
	walk(v, index, func(v reflect.Value) {
		start := len(b)

		if len(fields) > 1 {
			b = append(b, '(')
		}
		for i, f := range fields {
			if i > 0 {
				b = append(b, ", "...)
			}
			b = f.AppendValue(fmter, b, v)
		}
		if len(fields) > 1 {
			b = append(b, ')')
		}
		b = append(b, ", "...)

		if _, ok := seen[string(b[start:])]; ok {
			b = b[:start]
		} else {
			seen[string(b[start:])] = struct{}{}
		}
	})
	if len(seen) > 0 {
		b = b[:len(b)-2] // trim ", "
	}
	return b
}
