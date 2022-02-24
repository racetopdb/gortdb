// Go Rtdb Driver - A Rtdb-Driver for Go's database/sql pacakge.
package rtdb

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type rtdbConn struct {
	RtdbAdapter

	reset      bool // set for the sql/database/driver SessionResetter interface.
	config     *Config
	closed     AtomicBool
	closech    chan int
	ctxErr     AtomicError
	isWatching bool
}

func (rc *rtdbConn) deadline(ctx context.Context, now time.Time) time.Time {
	var earliest time.Time
	if rc.config.DialTimeout > 0 {
		earliest = now.Add(rc.config.DialTimeout)
	}
	if d, ok := ctx.Deadline(); ok {
		if earliest.IsZero() {
			return d
		}
		if d.IsZero() || earliest.Before(d) {
			return earliest
		}
	}
	return earliest
}

func (rc *rtdbConn) Begin() (driver.Tx, error) {
	panic("Do not support transaction!!!")
}

func (rc *rtdbConn) Close() (err error) {
	return rc.close()
}

func (rc *rtdbConn) Prepare(query string) (driver.Stmt, error) {
	panic("Prepared SQL Statement is not supported!!!!")
}

// Deprecated: Drivers should implement ExecerContext instead.
func (rc *rtdbConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if rc.closed.IsSet() {
		rtdbLogger.Println("err: rtdb is closed")
		return nil, driver.ErrBadConn
	}
	if len(args) != 0 {
		queryFmt, err := rc.formatArgs(query, args)
		if err != nil {
			return nil, err
		}
		query = queryFmt
	}

	if err := rc.exec(query); err != nil {
		return nil, err
	}

	return &rtdbResult{
		insertId:     int64(rc.insertId),
		affectedRows: int64(rc.affectedRows),
	}, nil
}

func (rc *rtdbConn) freeResult() error {
	if err := rc.freeResult(); err != nil {
		return err
	}
	return nil
}

func (rc *rtdbConn) query(query string, args []driver.Value) (*rtdbRows, error) {
	if rc.closed.IsSet() {
		rtdbLogger.Printf("before query, rtdb connection is closed")
		return nil, driver.ErrBadConn
	}
	if len(args) != 0 {
		queryFmt, err := rc.formatArgs(query, args)
		if err != nil {
			return nil, err
		}
		query = queryFmt
	}
	if DEBUG_PRINT_SQL {
		rtdbLogger.Printf("Query sql: %s\n", query)
	}
	// execute query
	if err := rc.CgoQuery(query, rc.config.Charset, rc.config.DBName); err != nil {
		return nil, err
	}
	// read result
	err := rc.ScanResult()
	if err != nil {
		return nil, err
	}
	if rc.IsResultSetEmpty() {
		return nil, nil
	}

	rows := &rtdbRows{
		rc: rc,
	}
	rows.resultSet.columns = rc.FetchFields()

	return rows, nil
}

// Deprecated: Drivers should implement QueryerContext instead.
func (rc *rtdbConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return rc.query(query, args)
}

func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dvalues := make([]driver.Value, len(named))

	for i, nameValue := range named {
		dvalues[i] = nameValue.Value
	}
	return dvalues, nil
}

func (rc *rtdbConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	var (
		result driver.Result
		err    error
	)
	values, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	if err = rc.withContext(ctx, func() error {
		var err error
		result, err = rc.Exec(query, values)
		return err
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (rc *rtdbConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	var (
		rows driver.Rows
		err  error
	)
	values, err := namedValueToValue(args)
	if err != nil {
		rtdbLogger.Println(err)
		return nil, err
	}
	if err = rc.withContext(ctx, func() error {
		var err error
		rows, err = rc.Query(query, values)
		return err
	}); err != nil {
		rtdbLogger.Println(err)
		return nil, err
	}
	return rows, nil
}

// Ping implements driver.Pinger interface.
func (rc *rtdbConn) Ping(ctx context.Context) (err error) {
	panic("Not implemented")
}

// BeginTx does not support database transaction.
func (rc *rtdbConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	panic("Do not support database transaction.")
}

func (rc *rtdbConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	panic("Prepared SQL Statement is not supported!!!!")
}

func (rc *rtdbConn) ResetSession(ctx context.Context) error {
	if rc.closed.IsSet() {
		rtdbLogger.Printf("err: rtdb is closed")
		return driver.ErrBadConn
	}
	rc.reset = true
	return nil
}

func (rc *rtdbConn) IsValid() bool {
	return rc.closed.IsSet()
}

func (rc *rtdbConn) formatArgs(query string, args []driver.Value) (string, error) {
	if strings.Count(query, "?") != len(args) {
		return "", driver.ErrSkip
	}
	var queryFmt string
	var argstrs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch v := arg.(type) {
		case int, uint8, int8, uint16, int16, uint32, int32, int64, uint64:
			argstrs = append(argstrs, fmt.Sprintf("%v", v))
		case float32, float64:
			// TODO: 可能需要优化
			argstrs = append(argstrs, fmt.Sprintf("%v", v))
		case bool:
			if v {
				argstrs = append(argstrs, "true")
			} else {
				argstrs = append(argstrs, "false")
			}
		case string:
			argstrs = append(argstrs, fmt.Sprintf("'%s'", v))
		case time.Time:
			if v.IsZero() {
				// TODO: handle zero time
			} else {
				timeStr := v.Format("2006-01-02 15:04:05.000")
				argstrs = append(argstrs, fmt.Sprintf("'%s'", timeStr))
			}
		case []byte:
			if v == nil {
				argstrs = append(argstrs, "NULL")
			} else {
				// TODO:
			}
		default:
			if v == nil {
				argstrs = append(argstrs, "NULL")
			} else {
				return "", driver.ErrSkip
			}
		}
	}
	old := query
	for i, arg := range argstrs {
		if i == 0 {
			old = query
		} else {
			old = queryFmt
		}
		queryFmt = strings.Replace(old, "?", arg, 1)
	}
	return queryFmt, nil
}

func (rc *rtdbConn) close() error {
	rc.closed.Set(true)
	close(rc.closech)
	if err := rc.CgoDisconnect(); err != nil {
		rtdbLogger.Printf("call c interface tsdb_disconnect failed, err: %v", err)
		return err
	}
	// finally clean up
	return nil
}

func (rc *rtdbConn) exec(query string) error {
	if DEBUG_PRINT_SQL {
		rtdbLogger.Printf("Exec sql: %s\n", query)
	}
	if err := rc.CgoQuery(query, rc.config.Charset, rc.config.DBName); err != nil {
		return err
	}

	if err := rc.CgoStoreResult(); err != nil {
		return err
	}
	return nil
}

func (rc *rtdbConn) error() error {
	if rc.closed.IsSet() {
		if err := rc.ctxErr.Error(); err != nil {
			return driver.ErrBadConn
		}
	}
	return nil
}

func (rc *rtdbConn) watchContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		rc.ctxErr.Set(ctx.Err())
		return err
	}
	if ctx.Done() == nil {
		return nil
	}
	// when ctx.Done returns a closed chan, ctx.Err() must return non-nil value.
	select {
	case <-ctx.Done():
		rc.ctxErr.Set(ctx.Err())
		return ctx.Err()
	case <-rc.closech:
	default:
	}
	return nil
}

func (rc *rtdbConn) withContext(ctx context.Context, f func() error) error {
	if err := rc.watchContext(ctx); err != nil {
		return err
	}
	if err := f(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		rc.ctxErr.Set(ctx.Err())
		return ctx.Err()
	case <-rc.closech:
	default:
		// it will always return
		if ctx.Done() == nil {
			return nil
		}
	}
	return nil
}
