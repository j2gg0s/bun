package pgdialect

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/j2gg0s/bun/dialect"
	"github.com/j2gg0s/bun/schema"
)

var (
	driverValuerType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

	stringType      = reflect.TypeOf((*string)(nil)).Elem()
	sliceStringType = reflect.TypeOf([]string(nil))

	intType      = reflect.TypeOf((*int)(nil)).Elem()
	sliceIntType = reflect.TypeOf([]int(nil))

	int64Type      = reflect.TypeOf((*int64)(nil)).Elem()
	sliceInt64Type = reflect.TypeOf([]int64(nil))

	float64Type      = reflect.TypeOf((*float64)(nil)).Elem()
	sliceFloat64Type = reflect.TypeOf([]float64(nil))
)

func appender(typ reflect.Type, pgArray bool) schema.AppenderFunc {
	switch typ.Kind() {
	case reflect.Uint32:
		return appendUint32ValueAsInt
	case reflect.Uint, reflect.Uint64:
		return appendUint64ValueAsInt
	case reflect.Ptr:
		return ptrAppenderFunc(typ, pgArray)
	case reflect.Slice:
		if pgArray {
			return arrayAppender(typ)
		}
	}
	return schema.Appender(typ)
}

func ptrAppenderFunc(typ reflect.Type, pgArray bool) schema.AppenderFunc {
	appender := appender(typ.Elem(), pgArray)
	return func(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
		if v.IsNil() {
			return dialect.AppendNull(b)
		}
		return appender(fmter, b, v.Elem())
	}
}

func appendUint32ValueAsInt(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
	return strconv.AppendInt(b, int64(int32(v.Uint())), 10)
}

func appendUint64ValueAsInt(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
	return strconv.AppendInt(b, int64(v.Uint()), 10)
}

//------------------------------------------------------------------------------

func arrayAppend(fmter schema.Formatter, b []byte, v interface{}) []byte {
	switch v := v.(type) {
	case int64:
		return strconv.AppendInt(b, v, 10)
	case float64:
		return dialect.AppendFloat64(b, v)
	case bool:
		return dialect.AppendBool(b, v)
	case []byte:
		return dialect.AppendBytes(b, v)
	case string:
		return arrayAppendString(b, v)
	case time.Time:
		return dialect.AppendTime(b, v)
	default:
		err := fmt.Errorf("pgdialect: can't append %T", v)
		return dialect.AppendError(b, err)
	}
}

func arrayElemAppender(typ reflect.Type) schema.AppenderFunc {
	if typ.Kind() == reflect.String {
		return arrayAppendStringValue
	}

	if typ.Implements(driverValuerType) {
		return arrayAppendDriverValue
	}

	return schema.Appender(typ)
}

func arrayAppendStringValue(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
	return arrayAppendString(b, v.String())
}

func arrayAppendDriverValue(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
	iface, err := v.Interface().(driver.Valuer).Value()
	if err != nil {
		return dialect.AppendError(b, err)
	}
	return arrayAppend(fmter, b, iface)
}

//------------------------------------------------------------------------------

func arrayAppender(typ reflect.Type) schema.AppenderFunc {
	kind := typ.Kind()
	if kind == reflect.Ptr {
		typ = typ.Elem()
		kind = typ.Kind()
	}

	switch kind {
	case reflect.Slice, reflect.Array:
		// ok:
	default:
		return nil
	}

	elemType := typ.Elem()

	if kind == reflect.Slice {
		switch elemType {
		case stringType:
			return appendStringSliceValue
		case intType:
			return appendIntSliceValue
		case int64Type:
			return appendInt64SliceValue
		case float64Type:
			return appendFloat64SliceValue
		}
	}

	appendElem := arrayElemAppender(elemType)
	if appendElem == nil {
		panic(fmt.Errorf("pgdialect: %s is not supported", typ))
	}

	return func(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
		kind := v.Kind()
		switch kind {
		case reflect.Ptr, reflect.Slice:
			if v.IsNil() {
				return dialect.AppendNull(b)
			}
		}

		if kind == reflect.Ptr {
			v = v.Elem()
		}

		b = append(b, '\'')

		b = append(b, '{')
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			b = appendElem(fmter, b, elem)
			b = append(b, ',')
		}
		if v.Len() > 0 {
			b[len(b)-1] = '}' // Replace trailing comma.
		} else {
			b = append(b, '}')
		}

		b = append(b, '\'')

		return b
	}
}

func appendStringSliceValue(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
	ss := v.Convert(sliceStringType).Interface().([]string)
	return appendStringSlice(b, ss)
}

func appendStringSlice(b []byte, ss []string) []byte {
	if ss == nil {
		return dialect.AppendNull(b)
	}

	b = append(b, '\'')

	b = append(b, '{')
	for _, s := range ss {
		b = arrayAppendString(b, s)
		b = append(b, ',')
	}
	if len(ss) > 0 {
		b[len(b)-1] = '}' // Replace trailing comma.
	} else {
		b = append(b, '}')
	}

	b = append(b, '\'')

	return b
}

func appendIntSliceValue(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
	ints := v.Convert(sliceIntType).Interface().([]int)
	return appendIntSlice(b, ints)
}

func appendIntSlice(b []byte, ints []int) []byte {
	if ints == nil {
		return dialect.AppendNull(b)
	}

	b = append(b, '\'')

	b = append(b, '{')
	for _, n := range ints {
		b = strconv.AppendInt(b, int64(n), 10)
		b = append(b, ',')
	}
	if len(ints) > 0 {
		b[len(b)-1] = '}' // Replace trailing comma.
	} else {
		b = append(b, '}')
	}

	b = append(b, '\'')

	return b
}

func appendInt64SliceValue(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
	ints := v.Convert(sliceInt64Type).Interface().([]int64)
	return appendInt64Slice(b, ints)
}

func appendInt64Slice(b []byte, ints []int64) []byte {
	if ints == nil {
		return dialect.AppendNull(b)
	}

	b = append(b, '\'')

	b = append(b, '{')
	for _, n := range ints {
		b = strconv.AppendInt(b, n, 10)
		b = append(b, ',')
	}
	if len(ints) > 0 {
		b[len(b)-1] = '}' // Replace trailing comma.
	} else {
		b = append(b, '}')
	}

	b = append(b, '\'')

	return b
}

func appendFloat64SliceValue(fmter schema.Formatter, b []byte, v reflect.Value) []byte {
	floats := v.Convert(sliceFloat64Type).Interface().([]float64)
	return appendFloat64Slice(b, floats)
}

func appendFloat64Slice(b []byte, floats []float64) []byte {
	if floats == nil {
		return dialect.AppendNull(b)
	}

	b = append(b, '\'')

	b = append(b, '{')
	for _, n := range floats {
		b = dialect.AppendFloat64(b, n)
		b = append(b, ',')
	}
	if len(floats) > 0 {
		b[len(b)-1] = '}' // Replace trailing comma.
	} else {
		b = append(b, '}')
	}

	b = append(b, '\'')

	return b
}

//------------------------------------------------------------------------------

func arrayAppendString(b []byte, s string) []byte {
	b = append(b, '"')
	for _, r := range s {
		switch r {
		case 0:
			// ignore
		case '\'':
			b = append(b, "'''"...)
		case '"':
			b = append(b, '\\', '"')
		case '\\':
			b = append(b, '\\', '\\')
		default:
			if r < utf8.RuneSelf {
				b = append(b, byte(r))
				break
			}
			l := len(b)
			if cap(b)-l < utf8.UTFMax {
				b = append(b, make([]byte, utf8.UTFMax)...)
			}
			n := utf8.EncodeRune(b[l:l+utf8.UTFMax], r)
			b = b[:l+n]
		}
	}
	b = append(b, '"')
	return b
}
