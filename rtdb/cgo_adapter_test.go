package rtdb_test

import (
	"github.com/racetopdb/gortdb/rtdb"
	"io"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	a *rtdb.RtdbAdapter
)

func TestMain(m *testing.M) {
	host := "192.168.1.43"
	user := "test"
	password := "test"
	port := 9000
	a = rtdb.NewRtdbAdapter(host, port, password, user)
	os.Exit(m.Run())
}

func Test_newRtdbAdapter(t *testing.T) {
	Convey("Test_newRtdbAdapter", t, func(ctx C) {
		var (
			host     string
			user     string
			password string
			port     int

			b *rtdb.RtdbAdapter
		)

		Convey("Then err should be nil", func(ctx C) {
			host = "192.168.1.43"
			user = "test"
			password = "test"
			port = 9000
			b = rtdb.NewRtdbAdapter(host, port, password, user)
			So(b, ShouldNotBeNil)
		})
	})
}

func TestRtdbAdapter(t *testing.T) {
	Convey("Test_newRtdbAdapter", t, func(ctx C) {
		var (
			err error
		)
		Convey("Test connect to database", func(ctx C) {
			err = a.CgoConnect()
			So(err, ShouldBeNil)
		})
		Convey("Test execute database query", func(ctx C) {
			err = a.CgoQuery("show databases;", "", "")
			So(err, ShouldBeNil)
		})
		Convey("Test store result set from database and then free result", func(ctx C) {
			err = a.CgoStoreResult()
			So(err, ShouldBeNil)
		})
	})
}

func mustExecuteSql(sql string, charset string, db string) {
	var err error
	if err = a.CgoConnect(); err != nil {
		panic(err)
	}
	if err = a.CgoQuery(sql, charset, db); err != nil {
		panic(err)
	}
}

// go test -timeout 30s -run ^TestRtdbAdapter_FetchFields$ github.com/racetopdb/gortdb/rtdb -v
func TestRtdbAdapter_FetchFields(t *testing.T) {
	Convey("Test fetch fields", t, func(ctx C) {
		mustExecuteSql("SHOW DATABASES;", "", "")
		Convey("Then err should be nil", func(ctx C) {
			fields := a.FetchFields()
			So(len(fields), ShouldEqual, 5)
		})
	})
}

// go test -timeout 30s -run ^TestRtdbAdapter_FetchOne$ github.com/racetopdb/gortdb/rtdb -v
func TestRtdbAdapter_FetchOne(t *testing.T) {
	Convey("Test fetch one row from stream", t, func(ctx C) {
		var (
			err error
		)
		mustExecuteSql("SHOW DATABASES;", "", "")
		fields := a.FetchFields()
		So(fields, ShouldNotBeEmpty)
		Convey("Then err should be nil", func(ctx C) {
			err = a.ScanResult()
			So(err, ShouldBeNil)
			row, err := a.FetchOne()
			So(err, ShouldBeNil)
			So(row, ShouldNotBeEmpty)
			spew.Config.Dump(row)
			// fetch again, should return io.EOF
			row, err = a.FetchOne()
			So(err, ShouldBeError, io.EOF)
			So(row, ShouldBeEmpty)
		})
	})
}

// go test -timeout 30s -run ^TestRtdbAdapter_C_KillMe$ github.com/racetopdb/gortdb/rtdb -v
func TestRtdbAdapter_C_KillMe(t *testing.T) {
	Convey("Test kill the rtdb client", t, func(ctx C) {
		var (
			err error
		)
		Convey("Then err should be nil", func(ctx C) {
			err = a.CgoKillMe()
			So(err, ShouldBeNil)
			// the rtdb client pointer will be nil
			// so it wll painc
			// So(func() {
			// 	mustExecuteSql("SHOW DATABASES;", "", "")
			// }, ShouldPanic)
		})
	})
}
