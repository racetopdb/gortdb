package rtdb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	rd     RtdbDriver
	dbTest *DBTest
)

var (
	user     string
	password string
	port     string
	address  string
	dsn      string
	dbname   string
)

type DBTest struct {
	db     *sql.DB
	logger *log.Logger
}

func (db *DBTest) mustExecute(sql string, args ...interface{}) sql.Result {
	res, err := db.db.Exec(sql, args...)
	if err != nil {
		db.Fatalf("exec", sql, err)
	}
	return res
}

func (db *DBTest) mustQuery(sql string, args ...interface{}) *sql.Rows {
	rows, err := db.db.Query(sql)
	if err != nil {
		db.Fatalf("query", sql, err)
	}
	return rows
}

func (db *DBTest) Fatalf(method string, query string, err error) {
	db.logger.Fatalf("[%s] executed, sql: %s catch err: %s", method, query, err)
}

func init() {
	user = getEnv("RTDB_TEST_USER", "test")
	password = getEnv("RTDB_TEST_PASSWORD", "test")
	port = getEnv("RTDB_TEST_PORT", "9000")
	address = getEnv("RTDB_TEST_ADDRESS", "127.0.0.1:9000")
	dbname = getEnv("RTDB_TEST_DBNAME", "test_db")
	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?param1=value1&param2=value2", user, password, address, dbname)
	_, err := rd.Open(dsn)
	if err != nil {
		panic(err)
	}
	db, err := sql.Open("rtdb", dsn)
	if err != nil {
		panic(err)
	}
	dbTest = &DBTest{
		db:     db,
		logger: log.New(os.Stdout, "[rtdb-test] ", log.Ldate|log.Lshortfile|log.Ltime|log.Ldate),
	}
}

// go test -timeout 30s -run ^TestCRUD$ github.com/racetopdb/gortdb/rtdb -v
func TestCRUD(t *testing.T) {
	// dbTest.mustExecute("CREATE TABLE test_table100(value bool)")

	Convey("Test CRUD", t, func(ctx C) {
		Convey("Test show databases;", func(ctx C) {
			type dbs struct {
				name        string
				blockSize   int
				bucketCount int
				tableCount  int
				path        string
			}
			d := &dbs{}
			rows := dbTest.mustQuery("SHOW DATABASES;")
			if rows.Next() {
				rows.Scan(&d.name, &d.blockSize, &d.bucketCount, &d.tableCount, &d.path)
				ShouldNotBeBlank(d.name)
				ShouldBeGreaterThan(d.blockSize, 0)
				ShouldNotBeBlank(d.path)
			}
		})

		Convey("Test insert data and then query", func(ctx C) {
			type testtable struct {
				time     time.Time
				isWoring bool
				age      int
				name     string
			}
			tt := &testtable{}
			dbTest.mustQuery("create database test_db if not exists;")
			dbTest.mustQuery("use test_db;")
			dbTest.mustQuery("create table if not exists test_table(is_working boolean, age int, name char(100));")
			dbTest.mustQuery("insert into test_table(is_working, age, name) values(false, 12, 'Michael');")
			rows := dbTest.mustQuery("select last * from test_table")
			if rows.Next() {
				rows.Scan(&tt.time, &tt.isWoring, &tt.age, &tt.name)
				spew.Dump(tt)
				ShouldBeTrue(tt.name == "Michael")
				ShouldBeTrue(tt.isWoring, false)
				ShouldBeTrue(tt.age, 12)
			}
		})
	})
}

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
