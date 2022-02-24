# gortdb - the go connector for rtdb

---

## Start
> {ProjectDirPath} 是项目的根目录路径

## Features

1. 轻量和快速
2. 使用CGO实现，依赖于C扩展；调用的C语言动态链接库libtsdb.so是rtdb数据库官方对外的纯c语言接口动态链接库，rtdbcli的go连接器通过CGO链接libtsdb.so来完成对rtdb数据库的操作
3. 支持Linux，在Windows下使用CGO编译默认只支持GCC工具链(MinGW或者Cygwin)，目前不支持在Windows下使用
4. 自带连接池(依赖于database/sql包实现)

## Requirements
* Go1.14或者更高版本
* github.com/davecgh/go-spew v1.1.1 调试打印数据的库
* github.com/smartystreets/goconvey v1.7.2 单元测试库
* 要使用CGO特性，在Linux上需要有GCC，同时需要确保CGO_ENABLED被设置为1
* 依赖于libtsdb.so的安装和对应的头文件(tsdb_ml.h位于{ProjectDirPath}/include目录下)

## Installation
* 开启Go mod支持
```shell
go env -w CGO_ENABLED=1
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.io,direct
```
* 下载gortdb包
```shell
go get github.com/racetopdb/gortdb
```

* 同步第三方包
```shell
go mod tidy
```
* 检查rtdb的C/C++连接器是否已经安装成功，Linux系统下查看/usr/lib/libtsdb.so是否存在；如果不存在，则需要使用本地的libtsdb.so。**如果使用本地的动态链接库(libtsdb.so，位于{ProjectDirPath}/dll/linux/libtsdb.do)，在启动程序编译之前必须要设置库检索路径，具体操作步骤如下**。

```shell
# 示例： export LD_LIBRARY_PATH=/home/zhangsan/source/gortdb/dll/linux
export LD_LIBRARY_PATH={ProjectDirPath}/dll/linux
```
## Usage

```Go
import (
	"github.com/racetopdb/gortdb/rtdb"
)

// mustQueryNoRows批量执行sql，不需要返回结果集
func (db *DBWrapper) mustQueryNoRows(sqls []string, args ...interface{}) {
	for _, sql := range sqls {
		_, err := db.db.Query(sql)
		if err != nil {
			db.Fatalf("query", sql, err)
		}
	}
}

func init() {
	// 设置为true，在执行sql之前会打印当前的sql
	rtdb.DEBUG_PRINT_SQL = true
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

/ ExampleTranscript
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
	start := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05.000")
	end := time.Now().AddDate(0, 0, 1).Format("2006-01-02 15:04:05.000")
	rows = db.mustQuery(fmt.Sprintf("select * from '%s' where time between '%s' and '%s'", tableName, start, end))
	for rows.Next() {
		rows.Scan(&studentTranscipts.time, &studentTranscipts.id, &studentTranscipts.studentName, &studentTranscipts.subjectNo, &studentTranscipts.subjectName, &studentTranscipts.score)
	}

	rows = db.mustQuery("select * from transcipt where time between ? and ?", time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1))
	for rows.Next() {
		rows.Scan(&studentTranscipts.time, &studentTranscipts.id, &studentTranscipts.studentName, &studentTranscipts.subjectNo, &studentTranscipts.subjectName, &studentTranscipts.score)
		fmt.Printf("Current student transcipts, name: %s, score: %d\n", studentTranscipts.studentName, studentTranscipts.score)
	}

	rows = db.mustQuery("select * from transcipt where time between ? and ? and student_name = ?", time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 1), "Faker")
	for rows.Next() {
		rows.Scan(&studentTranscipts.time, &studentTranscipts.id, &studentTranscipts.studentName, &studentTranscipts.subjectNo, &studentTranscipts.subjectName, &studentTranscipts.score)
		fmt.Printf("Current student transcipts, name: %s, score: %d\n", studentTranscipts.studentName, studentTranscipts.score)
	}
}
```

### dsn
dsn是连接数据库的参数字符串，默认格式"user:password@protocol(host:port)/dbname?parseTime=true&loc=Local"
* user 用户名，非必填
* password 用户密码，非必填
* protocol 连接的网络协议, 非必填，默认值是"tcp"
* host 主机地址, 非必填，默认值是"127.0.0.1"
* port 主机端口, 非必填，默认值是9000
* dbname 数据库名称, 非必填
* parseTime 是否解析时间， 非必填，默认值是True
* loc 时区，非必填，默认值是UTC

## API
```Go
// 通过一个数据库驱动和该驱动特定的数据源来打开数据库
// diverName: 数据库驱动名称
// dataSourceName: 数据库驱动特定的数据源名称
func Open(driverName, dataSourceName string) (*DB, error)

// Query执行一个返回数据库行数据的查询，通常是SELECT语句
// args用于查询中的任何占位符参数
func (db *DB) Query(query string, args ...interface{}) (*Rows, error)

// QueryContext执行一个带有上下文Context的查询，返回数据库行数据
// ctx: context上下文
// query: 查询SQL
// args: 用于查询中的任何占位符参数
func (rc *rtdbConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error)

// Exec执行一个不返回数据库行数据的查询
// query: 查询SQL
// args: 用于查询中的任何占位符参数
func (rc *rtdbConn) Exec(query string, args []driver.Value) (driver.Result, error)

// ExecContext执行一个不返回数据库行数据的查询, 用户可以传入上下
// ctx: context上下文
// query: 查询SQL
// args: 用于查询中的任何占位符参数
func (rc *rtdbConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error)
```