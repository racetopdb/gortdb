// Go Rtdb Driver - A Rtdb-Driver for Go's database/sql pacakge.
package rtdb

import (
	"context"
	"database/sql"
	"database/sql/driver"
)

type RtdbDriver struct{}

func (rd RtdbDriver) Open(dsn string) (driver.Conn, error) {
	config, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	c := &connector{
		config: config,
	}
	return c.Connect(context.Background())
}

func init() {
	sql.Register("rtdb", &RtdbDriver{})
}

func (rd RtdbDriver) OpenConnector(dsn string) (driver.Connector, error) {
	config, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	c := &connector{
		config: config,
	}
	return c, nil
}
