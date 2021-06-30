package example

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"
)

func NewWithMySQL() *bun.DB {
	sqldb, err := sql.Open("mysql", "bun_user:bun_password@tcp(localhost:31306)/bun")
	if err != nil {
		panic(err)
	}
	sqldb.SetMaxOpenConns(1)

	db := bun.NewDB(sqldb, mysqldialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose()))
	return db
}
