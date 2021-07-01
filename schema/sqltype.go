package schema

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/j2gg0s/bun/dialect"
	"github.com/j2gg0s/bun/dialect/sqltype"
	"github.com/j2gg0s/bun/internal"
)

var (
	bunNullTimeType = reflect.TypeOf((*NullTime)(nil)).Elem()
	nullTimeType    = reflect.TypeOf((*sql.NullTime)(nil)).Elem()
	nullBoolType    = reflect.TypeOf((*sql.NullBool)(nil)).Elem()
	nullFloatType   = reflect.TypeOf((*sql.NullFloat64)(nil)).Elem()
	nullIntType     = reflect.TypeOf((*sql.NullInt64)(nil)).Elem()
	nullStringType  = reflect.TypeOf((*sql.NullString)(nil)).Elem()
)

var sqlTypes = []string{
	reflect.Bool:          sqltype.Boolean,
	reflect.Int:           sqltype.BigInt,
	reflect.Int8:          sqltype.SmallInt,
	reflect.Int16:         sqltype.SmallInt,
	reflect.Int32:         sqltype.Integer,
	reflect.Int64:         sqltype.BigInt,
	reflect.Uint:          sqltype.BigInt,
	reflect.Uint8:         sqltype.SmallInt,
	reflect.Uint16:        sqltype.SmallInt,
	reflect.Uint32:        sqltype.Integer,
	reflect.Uint64:        sqltype.BigInt,
	reflect.Uintptr:       sqltype.BigInt,
	reflect.Float32:       sqltype.Real,
	reflect.Float64:       sqltype.DoublePrecision,
	reflect.Complex64:     "",
	reflect.Complex128:    "",
	reflect.Array:         "",
	reflect.Chan:          "",
	reflect.Func:          "",
	reflect.Interface:     "",
	reflect.Map:           sqltype.VarChar,
	reflect.Ptr:           "",
	reflect.Slice:         sqltype.VarChar,
	reflect.String:        sqltype.VarChar,
	reflect.Struct:        sqltype.VarChar,
	reflect.UnsafePointer: "",
}

func DiscoverSQLType(typ reflect.Type) string {
	switch typ {
	case timeType, nullTimeType, bunNullTimeType:
		return sqltype.Timestamp
	case nullBoolType:
		return sqltype.Boolean
	case nullFloatType:
		return sqltype.DoublePrecision
	case nullIntType:
		return sqltype.BigInt
	case nullStringType:
		return sqltype.VarChar
	}
	return sqlTypes[typ.Kind()]
}

//------------------------------------------------------------------------------

var jsonNull = []byte("null")

// NullTime is a time.Time wrapper that marshals zero time as JSON null and SQL NULL.
type NullTime struct {
	time.Time
}

var (
	_ json.Marshaler   = (*NullTime)(nil)
	_ json.Unmarshaler = (*NullTime)(nil)
	_ sql.Scanner      = (*NullTime)(nil)
	_ QueryAppender    = (*NullTime)(nil)
)

func (tm NullTime) MarshalJSON() ([]byte, error) {
	if tm.IsZero() {
		return jsonNull, nil
	}
	return tm.Time.MarshalJSON()
}

func (tm *NullTime) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, jsonNull) {
		tm.Time = time.Time{}
		return nil
	}
	return tm.Time.UnmarshalJSON(b)
}

func (tm NullTime) AppendQuery(fmter Formatter, b []byte) ([]byte, error) {
	if tm.IsZero() {
		return dialect.AppendNull(b), nil
	}
	return dialect.AppendTime(b, tm.Time), nil
}

func (tm *NullTime) Scan(src interface{}) error {
	if src == nil {
		tm.Time = time.Time{}
		return nil
	}

	switch src := src.(type) {
	case []byte:
		newtm, err := internal.ParseTime(internal.String(src))
		if err != nil {
			return err
		}

		tm.Time = newtm
		return nil
	case time.Time:
		tm.Time = src
		return nil
	default:
		return fmt.Errorf("bun: can't scan %#v into NullTime", src)
	}
}
