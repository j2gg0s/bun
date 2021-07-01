package main

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/j2gg0s/bun"
	"github.com/j2gg0s/bun/dialect/sqlitedialect"
	"github.com/j2gg0s/bun/extra/bundebug"
)

type Book struct {
	ID         int64
	Name       string
	CategoryID int64
}

var _ bun.AfterCreateTableHook = (*Book)(nil)

func (*Book) AfterCreateTable(ctx context.Context, query *bun.CreateTableQuery) error {
	_, err := query.DB().NewCreateIndex().
		Model((*Book)(nil)).
		Index("category_id_idx").
		Column("category_id").
		Exec(ctx)
	return err
}

func main() {
	ctx := context.Background()

	sqlite, err := sql.Open("sqlite3", ":memory:?cache=shared")
	if err != nil {
		panic(err)
	}
	sqlite.SetMaxOpenConns(1)

	db := bun.NewDB(sqlite, sqlitedialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose()))

	if _, err := db.NewCreateTable().Model((*Book)(nil)).Exec(ctx); err != nil {
		panic(err)
	}
}
