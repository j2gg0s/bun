module github.com/j2gg0s/bun/internal/dbtest

go 1.16

replace github.com/j2gg0s/bun => ../..

replace github.com/j2gg0s/bun/dbfixture => ../../dbfixture

replace github.com/j2gg0s/bun/dialect/pgdialect => ../../dialect/pgdialect

replace github.com/j2gg0s/bun/driver/pgdriver => ../../driver/pgdriver

replace github.com/j2gg0s/bun/driver/sqliteshim => ../../driver/sqliteshim

replace github.com/j2gg0s/bun/dialect/mysqldialect => ../../dialect/mysqldialect

replace github.com/j2gg0s/bun/dialect/sqlitedialect => ../../dialect/sqlitedialect

replace github.com/j2gg0s/bun/extra/bundebug => ../../extra/bundebug

require (
	github.com/bradleyjkemp/cupaloy v2.3.0+incompatible
	github.com/brianvoe/gofakeit/v6 v6.4.1
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.0.0
	github.com/jackc/pgx/v4 v4.11.0
	github.com/stretchr/testify v1.7.0
	github.com/j2gg0s/bun v0.1.1
	github.com/j2gg0s/bun/dbfixture v0.0.0-00010101000000-000000000000
	github.com/j2gg0s/bun/dialect/mysqldialect v0.1.0
	github.com/j2gg0s/bun/dialect/pgdialect v0.1.0
	github.com/j2gg0s/bun/dialect/sqlitedialect v0.1.0
	github.com/j2gg0s/bun/driver/pgdriver v0.1.0
	github.com/j2gg0s/bun/driver/sqliteshim v0.0.0-00010101000000-000000000000
	github.com/j2gg0s/bun/extra/bundebug v0.0.0-00010101000000-000000000000
	modernc.org/sqlite v1.11.1
)
