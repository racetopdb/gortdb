package rtdb

//
//#cgo LDFLAGS: -Wl,--allow-multiple-definition
//#cgo linux LDFLAGS: -L${SRCDIR}/../dll/linux -ltsdb
//#cgo linux CFLAGS: -I${SRCDIR}/../include
//#include "stdio.h"
//#include "stdlib.h"
//#include "tsdb_ml.h"
import "C"

// 使用系统安装路径
//#cgo linux CFLAGS: -I/usr/local/tsdb/include
//#cgo linux LDFLAGS: -L/usr/lib -ltsdb

// 使用本地绝对路径
//#cgo LDFLAGS: -L/home/pengzhao/rtdb/source/plugin/dynamiclib/linux -ltsdb
//#cgo linux CFLAGS: -I/home/pengzhao/rtdb/source/plugin/include

// 使用相对路径
// 这里查资料的时候理解错了，不管是库文件检索目录还是头文件检索目录都可以使用${SRCDIR}表示当前文件的绝对路径，然后再做相对路径的处理
//#cgo linux LDFLAGS: -L${SRCDIR}/../../dynamiclib/linux -ltsdb
//#cgo linux CFLAGS: -I${SRCDIR}/../../include

// 使用相对路径
// 在go项目内部的路径
//#cgo linux LDFLAGS: -L${SRCDIR}/../dll/linux -ltsdb
//#cgo linux CFLAGS: -I${SRCDIR}/../include

// export LD_LIBRARY_PATH=/home/pengzhao/rtdb/source/plugin/dynamiclib/linux

import (
	"database/sql/driver"
	"fmt"
	"io"
	"time"
	"unsafe"
)

const (
	MAX_FIELD_COUNT   uint32 = 1024
	VOID_POINTER_SIZE        = unsafe.Sizeof(unsafe.Pointer(nil))
)

const (
	rtdbAdapterStatusUnknown int16 = iota
	rtdbAdapterStatusDisconnect
	rtdbAdapterStatusConnected
	rtdbAdapterStatusFetchingResult
	rtdbAdapterStatusEOF
)

type FieldPtr *C.tsdb_v3_field_t
type MrFieldPtr *C.tsdb_ml_field_t
type ResultSetPtr *C.RTDB_RES_SET
type RowsPtr *C.tsdb_rows_t
type Row C.tsdb_row_t

type RtdbAdapter struct {
	dllPath      string
	connStr      string
	rtdbClient   unsafe.Pointer
	charset      string
	result       unsafe.Pointer
	fields       []rtdbField
	affectedRows uint64
	insertId     uint64
	cursor       RowsPtr // current row cursor, when read no rows, cursor will be nil.
	status       AtomicInt16
}

func NewRtdbAdapter(host string, port int, user string, password string) *RtdbAdapter {
	a := &RtdbAdapter{}
	a.connStr = buildConnStr(host, port, user, password)

	rtdbClient := unsafe.Pointer(C.tsdb_new())
	a.rtdbClient = rtdbClient
	a.charset = a.getCharset()
	return a
}

func buildConnStr(host string, port int, user string, password string) string {
	return fmt.Sprintf("user=%s;passwd=%s;servers=tcp://%s:%d", user, password, host, port)
}

func (a *RtdbAdapter) getStatus() int16 {
	status, ok := a.status.Get()
	if !ok {
		return rtdbAdapterStatusUnknown
	}
	return status
}

func (a *RtdbAdapter) setStatus(status int16) {
	a.status.Set(status)
}

func (a *RtdbAdapter) isConnected() bool {
	return a.getStatus() >= rtdbAdapterStatusConnected
}

// CgoConnect 使用Cgo调用C函数进行数据库连接
func (a *RtdbAdapter) CgoConnect() error {
	if a.isConnected() {
		rtdbLogger.Printf("Connection is not allowed in the current state, current state: %d\n", a.getStatus())
		return nil
	}
	cConnStr := C.CString(a.connStr)
	defer C.free(unsafe.Pointer(cConnStr))
	if err := convertErr(int(C.tsdb_connect(cConnStr))); err != nil {
		return err
	}
	a.setStatus(rtdbAdapterStatusConnected)
	return nil
}

// CgoDisconnect 使用Cgo调用C函数断开数据库连接
func (a *RtdbAdapter) CgoDisconnect() error {
	if !a.isConnected() {
		rtdbLogger.Printf("Connection destruction is not allowed in the current state, current state: %d\n", a.getStatus())
		return nil
	}
	if err := convertErr(int(C.tsdb_disconnect())); err != nil {
		return err
	}
	a.setStatus(rtdbAdapterStatusDisconnect)
	return nil
}

// CgoQuery 使用Cgo调用C函数执行一条数据库查询
func (a *RtdbAdapter) CgoQuery(sql string, charset string, db string) error {
	var charsetin string
	if charset == "" {
		charsetin = a.charset
	} else if a.charset != "" {
		charsetin = charset
	} else {
		a.charset = a.getCharset()
		charsetin = a.charset
	}

	cSql := C.CString(sql)
	cCharset := C.CString(charsetin)
	cDb := C.CString(db)
	defer C.free(unsafe.Pointer(cSql))
	defer C.free(unsafe.Pointer(cCharset))
	defer C.free(unsafe.Pointer(cDb))
	errCode := int(C.tsdb_query(a.rtdbClient, cSql, C.int(len(sql)), cCharset, cDb))
	if err := convertErr(errCode); err != nil {
		return err
	}
	return nil
}

func (a *RtdbAdapter) close() error {
	return a.CgoDisconnect()
}

func (a *RtdbAdapter) getCharset() string {
	cCharset := C.tsdb_charset_get()
	charset := C.GoString(cCharset)

	return charset
}

func (a *RtdbAdapter) readDone() bool {
	return a.getStatus() == rtdbAdapterStatusEOF
}

// CgoStoreResult 使用Cgo调用C函数获取查询的结果集
func (a *RtdbAdapter) CgoStoreResult() error {
	result := C.tsdb_store_result_v2(a.rtdbClient)
	if result != nil {
		a.result = unsafe.Pointer(result)
	}
	return nil
}

// CgoFreeResult 使用Cgo调用C函数释放查询结果集的内存
func (a *RtdbAdapter) CgoFreeResult() error {
	if err := convertErr(int(C.tsdb_free_result(a.rtdbClient, a.result))); err != nil {
		return err
	}
	return nil
}

// CleanUp 使用Cgo调用C函数进行最终的内存清理
func (a *RtdbAdapter) CleanUp() error {
	if a.rtdbClient == nil {
		return nil
	}
	if a.result != nil {
		// free result set
		if err := a.CgoFreeResult(); err != nil {
			return err
		}
	}

	if err := a.CgoKillMe(); err != nil {
		return err
	}

	return nil
}

func (a *RtdbAdapter) CgoKillMe() error {
	C.tsdb_kill_me(a.rtdbClient)
	return nil
}

// FetchFields 获取数据库列的信息
func (a *RtdbAdapter) FetchFields() []rtdbField {
	if len(a.fields) > 0 {
		return a.fields
	}
	var (
		f        *C.tsdb_ml_field_t
		i        int
		fields   []rtdbField
		fieldArr **C.tsdb_ml_field_t
	)
	fieldCount := C.int(0)
	fieldArr = C.tsdb_fetch_ml_fields(a.rtdbClient, &fieldCount)
	if unsafe.Pointer(fieldArr) == nil || int(fieldCount) <= 0 {
		return nil
	}
	for i = 0; i < int(fieldCount); i++ {
		field := rtdbField{}
		f = *(*MrFieldPtr)(unsafe.Pointer(uintptr(unsafe.Pointer(fieldArr)) + (unsafe.Sizeof(f) * uintptr(i))))
		if unsafe.Pointer(f) == nil || f == nil {
			break
		}
		name := C.GoString((*f).name)
		field.name = name
		field.fieldType = fieldType((*f).data_type)
		if uint8((*f).is_null) == 1 {
			field.isNull = true
		}
		field.len = uint8((*f).length)
		field.varLen = uint8((*f).real_length)
		fields = append(fields, field)
	}
	a.fields = fields
	return fields
}

func (a *RtdbAdapter) ScanResult() error {
	var (
		result ResultSetPtr
		rows   RowsPtr
	)

	result = C.tsdb_store_result_v2(a.rtdbClient)
	if unsafe.Pointer(result) == nil {
		// rtdbLogger.Printf("store result failed")
		// return NullPointer
		return nil
	}
	a.setStatus(rtdbAdapterStatusFetchingResult)
	rowCount := uint64((*result).row_count)
	a.result = unsafe.Pointer(result)
	a.affectedRows = rowCount
	if rowCount > 0 {
		rows = (*result).data
		a.cursor = rows
	}
	return nil
}

func (a *RtdbAdapter) IsResultSetEmpty() bool {
	return a.result == nil
}

// FetchOne 使用Cgo调用获取结果集的C函数，然后将C数据库结构转换为Go数据类型.
func (a *RtdbAdapter) FetchOne() (values []interface{}, err error) {
	if a.getStatus() != rtdbAdapterStatusFetchingResult {
		return nil, driver.ErrSkip
	}
	if a.getStatus() == rtdbAdapterStatusEOF {
		return nil, io.EOF
	}

	var (
		row        Row
		i          int
		fieldCount int
		value      interface{}
	)
	fieldCount = int((*(ResultSetPtr(a.result))).field_count)
	fields := a.fields
	row = Row((*a.cursor).row)
	for i = 0; i < fieldCount; i++ {
		fieldType := fields[i].fieldType
		switch fieldType {
		case fieldTypeUnknown:
			rtdbLogger.Printf("get fieldType is unknown, fieldIndex: %d, fieldName: %s", i, fields[i].name)
			continue

		case fieldTypeString:
			cString := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(row)) + (VOID_POINTER_SIZE * uintptr(i))))
			value = C.GoString(cString)
		case fieldTypeInt64:
			value = int64(*(*(**C.int64_t)(unsafe.Pointer(uintptr(unsafe.Pointer(row)) + (VOID_POINTER_SIZE * uintptr(i))))))
		case fieldTypeInt:
			value = int32(*(*(**C.int)(unsafe.Pointer(uintptr(unsafe.Pointer(row)) + (VOID_POINTER_SIZE * uintptr(i))))))
		case fieldTypeDouble:
			value = float64(*(*(**C.double)(unsafe.Pointer(uintptr(unsafe.Pointer(row)) + (VOID_POINTER_SIZE * uintptr(i))))))
		case fieldTypeFloat:
			value = float32(*(*(**C.float)(unsafe.Pointer(uintptr(unsafe.Pointer(row)) + (VOID_POINTER_SIZE * uintptr(i))))))
		case fieldTypeBinary:
			// TODO: handle binary field type
		case fieldTypeBool:
			tmp := byte(*(*(**C.byte_t)(unsafe.Pointer(uintptr(unsafe.Pointer(row)) + (VOID_POINTER_SIZE * uintptr(i))))))
			if tmp == 1 {
				value = true
			} else {
				value = false
			}
		case fieldTypeDatetime:
			v := int64(*(*(**C.int64_t)(unsafe.Pointer(uintptr(unsafe.Pointer(row)) + (VOID_POINTER_SIZE * uintptr(i))))))
			value = time.Unix(v/1000, 0)
		case fieldTypeNull:
			value = nil
		}
		values = append(values, value)
	}
	a.cursor = a.cursor.next
	if a.cursor == nil {
		a.setStatus(rtdbAdapterStatusEOF)
	}
	return
}

func convertErr(errCode int) error {
	noErrCode := 0
	switch errCode {
	case noErrCode:
		return nil
	case EINVAL:
		return InvalidArgs
	case EACCES:
		return NoAccess
	case ENOMEM:
		return OutOfMemory
	case EPROTO:
		return ProtocolError
	default:
		return ProtocolError
	}
}
