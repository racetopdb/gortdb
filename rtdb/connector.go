package rtdb

import (
	"context"
	"database/sql/driver"
)

type connector struct {
	config *Config
}

// TODO: how to use context for call c function (CgoConnect).
func (c *connector) Connect(cxt context.Context) (driver.Conn, error) {
	var (
		err error
	)
	host, port := c.config.HostAndPort()
	rc := &rtdbConn{
		RtdbAdapter: *NewRtdbAdapter(host, port, c.config.User, c.config.Password),
		config:      c.config,
		closech:     make(chan int),
	}

	if err = rc.withContext(cxt, func() error {
		return rc.CgoConnect()
	}); err != nil {
		return nil, err
	}

	return rc, nil
}

func (c *connector) Driver() driver.Driver {
	return &RtdbDriver{}
}
