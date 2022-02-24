package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/racetopdb/gortdb/rtdb"
)

type DBWrapper struct {
	db     *sql.DB
	logger *log.Logger
}

func (db *DBWrapper) mustQueryNoRows(sqls []string, args ...interface{}) {
	for _, sql := range sqls {
		_, err := db.db.Exec(sql)
		if err != nil {
			db.Fatalf("query", sql, err)
		}
	}
}

func (db *DBWrapper) mustQuery(sql string, args ...interface{}) *sql.Rows {
	rows, err := db.db.Query(sql, args...)
	if err != nil {
		db.Fatalf("query", sql, err)
	}
	return rows
}

func (db *DBWrapper) Fatalf(method, query string, err error) {
	db.logger.Fatalf("[%s] executed, sql: %s catch err: %s", method, query, err)
}

var (
	rd rtdb.RtdbDriver
	db *DBWrapper
)

var (
	user     string
	password string
	port     string
	address  string
	dsn      string
	dbname   string
)

func getEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func init() {
	user = getEnv("RTDB_TEST_USER", "test")
	password = getEnv("RTDB_TEST_PASSWORD", "test")
	port = getEnv("RTDB_TEST_PORT", "9000")
	address = getEnv("RTDB_TEST_ADDRESS", "127.0.0.1:9000")
	dbname = getEnv("RTDB_TEST_DBNAME", "test_db")
	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?param1=value1&param2=value2", user, password, address, dbname)
	// _, err := rd.Open(dsn)
	// if err != nil {
	// 	panic(err)
	// }
	sqlDB, err := sql.Open("rtdb", dsn)
	if err != nil {
		panic(err)
	}
	db = &DBWrapper{
		db:     sqlDB,
		logger: log.New(os.Stdout, "[rtdb] ", log.Ldate|log.Lshortfile|log.Ltime|log.Ldate),
	}
	sqls := []string{
		fmt.Sprintf("CREATE DATABASE '%s' IF NOT EXISTS;", dbname),
		fmt.Sprintf("USE '%s';", dbname),
	}
	db.mustQueryNoRows(sqls)
}

// ExampleTranscript
// 流程步骤：
// 1- 创建成绩单表transcipts(id int, student_name char(100), subject_no int, subject_name char(100), score int)
// id表示学生id, student_name表示学生姓名，subject_no表示课程号，subject_name表示课程名称，score表示学生分数。
// 2- 随机插入100条数据。
// 3- 查询出最后一条数据并打印。
// 4- 按照时间范围进行查询。
func ExampleTranscript() {
	type transcipt struct {
		time        time.Time
		id          int
		studentName string
		subjectNo   int
		subjectName string
		score       int
	}
	transcipts := func(count int) []*transcipt {
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s)
		subjectNames := []string{
			"Chinese",
			"Math",
			"English",
			"Chemical",
			"Physics",
		}
		studentNames := []string{
			"Michael",
			"Jane",
			"Xiaokang",
			"Uzi",
			"Faker",
			"Xiye",
			"Theshy",
			"Rookie",
			"JackieLove",
			"Tian",
		}
		var ret []*transcipt
		for i := 0; i < count; i++ {
			t := &transcipt{
				subjectName: subjectNames[r.Intn(len(subjectNames))],
				subjectNo:   r.Intn(len(subjectNames) + 1),
				score:       r.Intn(40) + 60,
				studentName: studentNames[r.Intn(len(studentNames))],
			}
			ret = append(ret, t)
		}
		return ret
	}(100)
	tableName := "transcipt"
	sqls := []string{
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS '%s'(id int, student_name char(100), subject_no int, subject_name char(100), score int)", tableName),
	}
	// 随机插入数据
	db.mustQueryNoRows(sqls)
	var insertSqls []string
	for i := 0; i < 100; i++ {
		t := transcipts[i]
		sql := fmt.Sprintf("INSERT INTO '%s'(id, student_name, subject_no, subject_name, score) VALUES(%d, '%s', %d, '%s', %d)", tableName, i, t.studentName, t.subjectNo, t.subjectName, t.score)
		insertSqls = append(insertSqls, sql)
	}
	db.mustQueryNoRows(insertSqls)

	// 查询最后一条数据
	rows := db.mustQuery(fmt.Sprintf("SELECT LAST * FROM '%s'", tableName))
	studentTranscipts := transcipt{}
	for rows.Next() {
		rows.Scan(&studentTranscipts.time, &studentTranscipts.id, &studentTranscipts.studentName, &studentTranscipts.subjectNo, &studentTranscipts.subjectName, &studentTranscipts.score)
		fmt.Printf("Current student transcipts, name: %s, score: %d\n", studentTranscipts.studentName, studentTranscipts.score)
	}

	// 通过时间范围查询数据
	rows = db.mustQuery("select * from transcipt where time between ? and ?", time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	for rows.Next() {
		rows.Scan(&studentTranscipts.time, &studentTranscipts.id, &studentTranscipts.studentName, &studentTranscipts.subjectNo, &studentTranscipts.subjectName, &studentTranscipts.score)
		fmt.Printf("Current student transcipts, time: %s, name: %s, score: %d\n", studentTranscipts.time.Format("2006-01-02 15:04:05.999"), studentTranscipts.studentName, studentTranscipts.score)
	}

	rows = db.mustQuery("select * from transcipt where time between ? and ? and student_name = ?", time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1), "Faker")
	for rows.Next() {
		rows.Scan(&studentTranscipts.time, &studentTranscipts.id, &studentTranscipts.studentName, &studentTranscipts.subjectNo, &studentTranscipts.subjectName, &studentTranscipts.score)
		fmt.Printf("Current student transcipts, time: %s, name: %s, score: %d\n", studentTranscipts.time.Format("2006-01-02 15:04:05.999"), studentTranscipts.studentName, studentTranscipts.score)
	}
}

func main() {
	ExampleTranscript()
}
