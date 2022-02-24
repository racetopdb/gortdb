package rtdb

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	. "github.com/smartystreets/goconvey/convey"
)

// go test -timeout 30s -run ^TestParseDSN$ github.com/racetopdb/gortdb/rtdb -v
func TestParseDSN(t *testing.T) {
	Convey("TestParseDSN", t, func(ctx C) {
		var (
			dsn    string
			err    error
			config *Config
		)
		Convey("Then err should be nil", func(ctx C) {
			dsn = "user:password@protocol(host:port)/dbname?parseTime=true&loc=Local"
			config, err = ParseDSN(dsn)
			So(err, ShouldBeNil)
			So(config, ShouldNotBeNil)
			spew.Dump(config)
		})

		Convey("Test the correct dsn and then err should be nil", func(ctx C) {
			var testDSNs = []struct {
				param  string
				result *Config
			}{
				{"user:password@protocol(host:port)/dbname?param1=value1&param2=value2",
					&Config{User: "user", Password: "password", Protocol: "protocol", Address: "host:port", DBName: "dbname", Charset: "iso-8859-1", Location: time.UTC, DialTimeout: time.Millisecond * 500, Params: map[string]string{
						"param1": "value1",
						"param2": "value2",
					}}},
				{"root:asd123456@tcp(127.0.0.1:9000)/myDB?parseTime=true&charset=UTF-8&loc=UTC",
					&Config{User: "root", Password: "asd123456", Protocol: "tcp", ParseTime: true, Address: "127.0.0.1:9000", DBName: "myDB", Charset: "utf-8", Location: time.UTC, DialTimeout: time.Millisecond * 500, Params: map[string]string{
						"parseTime": "true",
						"charset":   "UTF-8",
						"loc":       "UTC",
					}}},
				{"root@unix(/path/to/socket)/myDB?charset=UTF-8",
					&Config{User: "root", Protocol: "unix", Address: "/path/to/socket", DBName: "myDB", Charset: "utf-8", Location: time.UTC, DialTimeout: time.Millisecond * 500, Params: map[string]string{
						"charset": "UTF-8",
					}}},
				{
					"/dbname",
					&Config{DBName: "dbname", Charset: "iso-8859-1", Location: time.UTC, DialTimeout: time.Millisecond * 500, Protocol: "tcp", Address: "127.0.0.1:9000"},
				},
				{
					"/dbname?parseTime=true&charset=UTF-8&loc=UTC",
					&Config{DBName: "dbname", Charset: "utf-8", Location: time.UTC, DialTimeout: time.Millisecond * 500, Params: map[string]string{
						"parseTime": "true",
						"charset":   "UTF-8",
						"loc":       "UTC"},
						ParseTime: true,
						Protocol:  "tcp", Address: "127.0.0.1:9000",
					},
				},
			}
			for _, testDSN := range testDSNs {
				config, err = ParseDSN(testDSN.param)
				So(err, ShouldBeNil)
				// ShouldResemble receives exactly two parameters and does a deep equal check (see reflect.DeepEqual)
				So(config, ShouldResemble, testDSN.result)
			}
		})
	})
}
