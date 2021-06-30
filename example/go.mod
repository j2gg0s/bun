module github.com/uptrace/bun/example

go 1.16

replace github.com/uptrace/bun => ../

replace github.com/uptrace/bun/extra/bundebug => ../extra/bundebug

replace github.com/uptrace/bun/dialect/mysqldialect => ../dialect/mysqldialect

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/uptrace/bun v0.2.12
	github.com/uptrace/bun/dialect/mysqldialect v0.2.12
	github.com/uptrace/bun/extra/bundebug v0.2.12
)
