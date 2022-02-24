// Go Rtdb Driver - A Rtdb-Driver for Go's database/sql pacakge.
package rtdb

import "database/sql/driver"

type rtdbStmt struct{}

func (s *rtdbStmt) Close() error {
	panic("not implemented")
}

func (s *rtdbStmt) NumInput() int {
	panic("not implemented")
}

func (s *rtdbStmt) Exec(args driver.Value) (driver.Result, error) {
	panic("not implemented")
}

func (s *rtdbStmt) Query(args []driver.Value) (driver.Rows, error) {
	panic("not implemented")
}
