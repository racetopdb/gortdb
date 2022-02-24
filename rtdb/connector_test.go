package rtdb

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// go test -timeout 30s -run ^Test_connector_Connect$ github.com/racetopdb/gortdb/rtdb -v
func Test_connector_Connect(t *testing.T) {
	Convey("Test_connector_Connect with multiple context", t, func(ctx C) {
		var (
			baseCtx context.Context
			cancel  context.CancelFunc
			err     error
		)
		connector := &connector{
			config: &Config{
				User:        "test",
				Password:    "test",
				Protocol:    "tcp",
				Address:     "192.168.1.43:9000",
				DBName:      "TEST_DB",
				DialTimeout: time.Second * 10,
			},
		}
		Convey("Test_connector_Connect with cancel context", func(ctx C) {
			baseCtx, cancel = context.WithCancel(context.Background())
			defer cancel()
			_, err = connector.Connect(baseCtx)
			So(err, ShouldBeNil)

			baseCtx, cancel = context.WithCancel(context.Background())
			cancel()
			_, err = connector.Connect(baseCtx)
			So(err, ShouldBeError, context.Canceled)
		})
		Convey("Test_connector_Connect with deadline context", func(ctx C) {
			baseCtx, cancel = context.WithDeadline(context.Background(), time.Now().Add(connector.config.DialTimeout))
			defer cancel()
			_, err = connector.Connect(baseCtx)
			So(err, ShouldBeNil)

			// it will time out
			baseCtx, cancel = context.WithDeadline(context.Background(), time.Now().Add(time.Nanosecond*10))
			defer cancel()
			_, err = connector.Connect(baseCtx)
			So(err, ShouldBeError, context.DeadlineExceeded)

			// it will cancel
			baseCtx, cancel = context.WithDeadline(context.Background(), time.Now().Add(time.Microsecond*10))
			cancel()
			_, err = connector.Connect(baseCtx)
			So(err, ShouldBeError, context.Canceled)
		})
	})
}
