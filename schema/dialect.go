package schema

import (
	"reflect"
	"sync"

	"github.com/j2gg0s/bun/dialect/feature"
)

type Dialect interface {
	Name() string
	Features() feature.Feature

	Tables() *Tables
	OnTable(table *Table)

	IdentQuote() byte
	Append(fmter Formatter, b []byte, v interface{}) []byte
	Appender(typ reflect.Type) AppenderFunc
	Scanner(typ reflect.Type) ScannerFunc
}

//------------------------------------------------------------------------------

type nopDialect struct {
	tables   *Tables
	features feature.Feature

	appenderMap sync.Map
	scannerMap  sync.Map
}

func newNopDialect() *nopDialect {
	d := new(nopDialect)
	d.tables = NewTables(d)
	d.features = feature.Returning
	return d
}

func (d *nopDialect) Name() string {
	return ""
}

func (d *nopDialect) Features() feature.Feature {
	return d.features
}

func (d *nopDialect) Tables() *Tables {
	return d.tables
}

func (d *nopDialect) OnField(field *Field) {}

func (d *nopDialect) OnTable(table *Table) {}

func (d *nopDialect) IdentQuote() byte {
	return '"'
}

func (d *nopDialect) Append(fmter Formatter, b []byte, v interface{}) []byte {
	return Append(fmter, b, v)
}

func (d *nopDialect) Appender(typ reflect.Type) AppenderFunc {
	if v, ok := d.appenderMap.Load(typ); ok {
		return v.(AppenderFunc)
	}

	fn := Appender(typ)

	if v, ok := d.appenderMap.LoadOrStore(typ, fn); ok {
		return v.(AppenderFunc)
	}
	return fn
}

func (d *nopDialect) Scanner(typ reflect.Type) ScannerFunc {
	if v, ok := d.scannerMap.Load(typ); ok {
		return v.(ScannerFunc)
	}

	fn := Scanner(typ)

	if v, ok := d.scannerMap.LoadOrStore(typ, fn); ok {
		return v.(ScannerFunc)
	}
	return fn
}
