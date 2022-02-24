package rtdb

import (
	"database/sql/driver"
	"io"
	"reflect"
)

type rtdbResult struct {
	insertId     int64
	affectedRows int64
}

// LastInsertId returns the database's auto-generated ID
// after, for example, an INSERT into a table with primary
// key.When execute a SELECT, return zero and nil.
func (r *rtdbResult) LastInsertId() (int64, error) {
	return r.insertId, nil
}

// RowsAffected returns the number of rows affected by the
// query.
func (r *rtdbResult) RowsAffected() (int64, error) {
	return r.affectedRows, nil
}

type rtdbRows struct {
	rc        *rtdbConn
	resultSet rtdbResultSet
}

func (r *rtdbRows) Columns() []string {
	if r.resultSet.columnNames != nil {
		return r.resultSet.columnNames
	}

	columns := make([]string, len(r.resultSet.columns))
	for i := range columns {
		columns[i] = r.resultSet.columns[i].name
	}
	r.resultSet.columnNames = columns
	return columns
}

func (r *rtdbRows) Close() error {
	return nil
}

func (r *rtdbRows) Next(dest []driver.Value) error {
	if rc := r.rc; rc != nil {
		if err := rc.error(); err != nil {
			return err
		}
	}
	return r.fetchOne(dest)
}

// HasNextResultSet always return false because rtdb dose not support multiple result set.
func (r *rtdbRows) HasNextResultSet() bool {
	return false
}

// NextResultSet always return nil because rtdb dose not support multiple result set.
func (r *rtdbRows) NextResultSet() error {
	return nil
}

// fetchOne fetch one row from rtdb connection.
func (r *rtdbRows) fetchOne(dest []driver.Value) error {
	rc := r.rc
	if r.readDone() {
		return io.EOF
	}
	values, err := rc.FetchOne()
	if err != nil {
		return err
	}
	if len(values) != len(dest) {
		return driver.ErrSkip
	}
	for i := range dest {
		dest[i] = driver.Value(values[i])
	}
	return nil

}

func (r *rtdbRows) readDone() bool {
	return r.rc.readDone()
}

func (r *rtdbRows) ColumnTypeDatabaseTypeName(i int) string {
	return r.resultSet.columns[i].typeDatabaseTypeName()
}

func (r *rtdbRows) ColumnTypeScanType(i int) reflect.Type {
	return r.resultSet.columns[i].scanType()
}

func (r *rtdbRows) ColumnTypePrecisionScale(i int) (int64, int64, bool) {
	return -1, -1, false
}

type rtdbResultSet struct {
	columns     []rtdbField
	columnNames []string
}
