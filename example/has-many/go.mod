module github.com/j2gg0s/bun/example/has-many

go 1.16

replace github.com/j2gg0s/bun => ../..

replace github.com/j2gg0s/bun/extra/bundebug => ../../extra/bundebug

replace github.com/j2gg0s/bun/dialect/sqlitedialect => ../../dialect/sqlitedialect

require (
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/j2gg0s/bun v0.1.1
	github.com/j2gg0s/bun/dialect/sqlitedialect v0.0.0-20210507070510-0d95488a5553
	github.com/j2gg0s/bun/extra/bundebug v0.0.0-00010101000000-000000000000
)
