package rtdb

import (
	"database/sql/driver"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	rc = &rtdbConn{}
)

func Test_formatArgs(t *testing.T) {

	var (
		query    string
		args     []driver.Value
		queryfmt string
		err      error
	)
	Convey("Test_formatArgs", t, func(ctx C) {
		Convey("Then err should be nil", func(ctx C) {
			query = "SELECT * from table_foo where name = ?"
			args = []driver.Value{"bar"}

			queryfmt, err = rc.formatArgs(query, args)
			So(err, ShouldBeNil)
			So(queryfmt, ShouldNotBeBlank)
		})
	})
}
