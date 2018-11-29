package main

import (
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

func checkErr1(err error) {
	if err != nil {
		panic(err)
	}
}

type inserter interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

func insert(i inserter, s string) error {
	for {
		_, err := i.Exec("INSERT INTO foo (name) VALUES (?)", s)
		if err == nil {
			return nil
		}
		if err == sqlite3.ErrLocked || err == sqlite3.ErrBusy || err.Error() == "database table is locked" {
			continue
		}
		return err
	}
}

//func main() {
//	db, err := sql.Open("sqlite3", ":memory:")
//	if err != nil {
//		panic(err)
//	}
//
//	_, err = db.Exec("CREATE TABLE foo (id INTEGER NOT NULL PRIMARY KEY, name TEXT)", nil)
//	if err != nil {
//		panic(err)
//	}
//
//	err = insert(db, "hello")
//	if err != nil {
//		panic(err)
//	}
//	db.Close()
//}

type temperatureData1 struct {
	id            int
	time          int
	timeZone      int
	algVer        string
	temperature   int
	stable        int
	wearScore     int
	wearGrade     int
	source        int
	sn            string
	hardwareVer   string
	firmwareVer   string
	patientId     string
	usageScenario int
	isUpload      int
}

func main01() {
	db, err := sql.Open("sqlite3", "X:/Golang/mysql/TemperatureDbForERAS.db")
	//db, err := sql.Open("sqlite3", "./1.db")
	checkErr1(err)

	fmt.Println("打开数据库成功")

	//rows, err := db.Query("SELECT name FROM new_table WHERE sex = ?", 27)
	rows, err := db.Query("SELECT * FROM body_temperature ORDER BY time ASC")
	if err != nil {
		log.Fatal(err)
	}
	column, err := rows.Columns()
	checkErr1(err)
	for _, name := range column {
		fmt.Print(name)
		fmt.Print(" ")
	}
	fmt.Println()
	//id            int
	//time          int
	//timeZone      int
	//algVer        string
	//temperature   int
	//stable        int
	//wearScore     int
	//wearGrade     int
	//source        int
	//sn            string
	//hardwareVer   string
	//firmwareVer   string
	//patientId     string
	//usageScenario int
	//isUpload      int
	var data temperatureData1
	for rows.Next() {
		err = rows.Scan(&data.id, &data.time, &data.timeZone, &data.algVer, &data.temperature, &data.stable, &data.wearScore,
			&data.wearGrade, &data.source, &data.sn, &data.hardwareVer, &data.firmwareVer, &data.patientId, &data.usageScenario, &data.isUpload)
		checkErr1(err)
		strTime := time.Unix(int64(data.time), 0).Format("2006-01-02 15:04:05")
		//strTime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Print(strTime)
		fmt.Println(data)
	}

	////插入数据
	//stmt, err := db.Prepare("INSERT INTO userinfo(username, departname, created) values(?,?,?)")
	//checkErr(err)
	//
	//res, err := stmt.Exec("astaxie", "研发部门", "2012-12-09")
	//checkErr(err)
	//
	//id, err := res.LastInsertId()
	//checkErr(err)
	//
	//fmt.Println(id)
	////更新数据
	//stmt, err = db.Prepare("update userinfo set username=? where uid=?")
	//checkErr(err)
	//
	//res, err = stmt.Exec("astaxieupdate", id)
	//checkErr(err)
	//
	//affect, err := res.RowsAffected()
	//checkErr(err)
	//
	//fmt.Println(affect)
	//
	////查询数据
	//rows, err := db.Query("SELECT * FROM userinfo")
	//checkErr(err)
	//
	//for rows.Next() {
	//	var uid int
	//	var username string
	//	var department string
	//	var created string
	//	err = rows.Scan(&uid, &username, &department, &created)
	//	checkErr(err)
	//	fmt.Println(uid)
	//	fmt.Println(username)
	//	fmt.Println(department)
	//	fmt.Println(created)
	//}
	//
	////删除数据
	//stmt, err = db.Prepare("delete from userinfo where uid=?")
	//checkErr(err)
	//
	//res, err = stmt.Exec(id)
	//checkErr(err)
	//
	//affect, err = res.RowsAffected()
	//checkErr(err)
	//
	//fmt.Println(affect)
	db.Close()
}

//func checkErr(err error) {
//	if err != nil {
//		panic(err)
//	}
//}
