module github.com/uptrace/bun/example/basic

go 1.16

replace github.com/uptrace/bun/dbfixture => ../../dbfixture

replace github.com/uptrace/bun/example => ../

replace github.com/uptrace/bun => ../..

replace github.com/uptrace/bun/extra/bundebug => ../../extra/bundebug

replace github.com/uptrace/bun/dialect/mysqldialect => ../../dialect/mysqldialect

require (
	github.com/uptrace/bun/dbfixture v0.0.0-00010101000000-000000000000
	github.com/uptrace/bun/example v0.0.0-00010101000000-000000000000
)
