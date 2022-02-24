package rtdb

import (
	"database/sql"
	"reflect"
)

type fieldType uint8

const (
	fieldTypeUnknown fieldType = iota
	fieldTypeBool
	fieldTypeInt
	fieldTypeInt64
	fieldTypeFloat
	fieldTypeDouble
	fieldTypeBinary
	fieldTypeString
	fieldTypeDatetime
	fieldTypeNull
)

var (
	scanTypeUnknown  = reflect.TypeOf(new(interface{}))
	scanTypeFloat32  = reflect.TypeOf(float32(0))
	scanTypeFloat64  = reflect.TypeOf(float64(0))
	scanTypeInt8     = reflect.TypeOf(int8(0))
	scanTypeUint8    = reflect.TypeOf(uint8(0))
	scanTypeInt16    = reflect.TypeOf(int16(0))
	scanTypeUint16   = reflect.TypeOf(uint16(0))
	scanTypeInt32    = reflect.TypeOf(int32(0))
	scanTypeUint32   = reflect.TypeOf(uint32(0))
	scanTypeInt64    = reflect.TypeOf(int64(0))
	scanTypeUint64   = reflect.TypeOf(uint64(0))
	scanTypeInt      = reflect.TypeOf(int(0))
	scanTypeBool     = reflect.TypeOf(bool(false))
	scanTypeRawBytes = reflect.TypeOf(sql.RawBytes{})

	scanTypeNullFloat = reflect.TypeOf(sql.NullFloat64{})
	scanTypeNullInt   = reflect.TypeOf(sql.NullInt64{})
	scanTypeNullTime  = reflect.TypeOf(sql.NullTime{})
	scanTypeNullBool  = reflect.TypeOf(sql.NullBool{})
)

type rtdbField struct {
	name      string
	len       uint8
	fieldType fieldType
	charset   uint8
	isNull    bool
	varLen    uint8 // for variable length data structure
}

func (rf *rtdbField) scanType() reflect.Type {
	switch rf.fieldType {
	case fieldTypeUnknown:
		return scanTypeUnknown
	case fieldTypeBool:
		if rf.isNull {
			return scanTypeNullBool
		}
		return scanTypeBool
	case fieldTypeInt:
		if rf.isNull {
			return scanTypeNullInt
		}
		return scanTypeInt
	case fieldTypeInt64:
		if rf.isNull {
			return scanTypeNullInt
		}
		return scanTypeInt64
	case fieldTypeFloat:
		if rf.isNull {
			return scanTypeNullFloat
		}
		return scanTypeFloat32
	case fieldTypeDouble:
		if rf.isNull {
			return scanTypeNullFloat
		}
		return scanTypeFloat64
	case fieldTypeBinary, fieldTypeString:
		return scanTypeRawBytes
	default:
		return scanTypeUnknown
	}
}

func (rf *rtdbField) typeDatabaseTypeName() string {
	switch rf.fieldType {
	case fieldTypeBool:
		return "BOOL"
	case fieldTypeDatetime:
		return "DATETIME"
	case fieldTypeFloat:
		return "FLOAT"
	case fieldTypeDouble:
		return "DOUBLE"
	case fieldTypeString:
		return "STRING"
	case fieldTypeInt:
		return "INT"
	case fieldTypeInt64:
		return "INT64"
	case fieldTypeNull:
		return "NULL"
	default:
		return ""
	}
}
