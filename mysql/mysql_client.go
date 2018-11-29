package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strconv"
	"time"
)

type DbWorker struct {
	//mysql data source name
	Dsn string
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

const (
	TEMP_TABLE_NAME      = "raiing_tcms_temp_data"
	B2W_TABLE_NAME       = "raiing_tcms_b2w"
	USER_TABLE_NAME      = "raiing_tcms_user"
	EVENT_TABLE_NAME     = "raiing_tcms_event_data"
	TEMPERATURE_INTERVAL = 4 // 温度间隔
)

type raiingTCMSUser struct {
	id             int64
	uuid           string
	caseNum        string // 病人ID
	bedNum         int
	name           int
	sex            int
	birthday       string
	model          int
	height         int
	weight         int
	inHospitalTime int64
	pacing         int
	hospital       string
	department     string
	status         int
	addTime        int64
	addId          int
	lastUpdateTime int64
	lastUpdateId   int64
}

type raiingTCMSTempData struct {
	id               int
	uuid             string
	userUuid         string
	time             int64
	timeZone         int
	hardwareSn       string
	hardwareVersion  string
	firmwareVersion  string
	b2wSn            string
	b2wVersion       string
	algorithmVersion string
	appVersion       string
	dataVersion      string
	tempMin          int
	tempMax          int
	tempAvg          int
	tempValid        int
	addTime          int64
	lastUpdateTime   int64
	data             string
}

//{"time":1540859820,"temp":36842,"wear_quality":81,"stable":3,"wear_stage":1,"add_time":1540859881}
type temperatureData struct {
	Time        int64 `json:"time"`
	Temp        int   `json:"temp"`
	WearQuality int   `json:"wear_quality"`
	Stable      int   `json:"stable"`
	WearStage   int   `json:"wear_stage"`
	AddTime     int64 `json:"add_time"`
}

type raiingTCMSEventData struct {
	id             int
	uuid           string
	casesId        string // 病人ID
	userUuid       string
	eventUuid      string
	eventType      int
	timezone       int
	startTime      int64
	endTime        int64
	createTime     int64
	updateTime     int64
	detail         string
	addTime        int64
	lastUpdateTime int64
}

func main() {
	dbw := DbWorker{
		Dsn: "root:123456@tcp(127.0.0.1:3306)/raiing_tcms_v6_temp",
	}
	db, err := sql.Open("mysql",
		dbw.Dsn)
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("数据库打开成功！")
	defer db.Close()
	//rows1, err := db.Query("SELECT b2w_sn, hardware_sn FROM " + TEMP_TABLE_NAME)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//b2wSNs := make(map[string]interface{}, 10)
	//hardwareSNs := make(map[string]interface{}, 10)
	//for rows1.Next() {
	//	var b2wSN string
	//	var hardwareSN string
	//	err := rows1.Scan(&b2wSN, &hardwareSN)
	//	checkErr(err)
	//	b2wSNs[strings.Trim(b2wSN, " ")] = "_"
	//	hardwareSNs[strings.Trim(hardwareSN, " ")] = "_"
	//}
	//fmt.Println("705序列号: ", len(hardwareSNs), hardwareSNs)
	//fmt.Println("B2W序列号: ", len(b2wSNs), b2wSNs)

	rows3, err := db.Query("SELECT uuid, case_num FROM " + USER_TABLE_NAME)
	if err != nil {
		log.Fatal(err)
	}
	userUUIDS := make(map[string]string, 10) // 用户UUID
	for rows3.Next() {
		var userUUID string
		var caseNum string
		err := rows3.Scan(&userUUID, &caseNum)
		checkErr(err)
		userUUIDS[userUUID] = caseNum
	}
	fmt.Println("用户UUID: ", len(userUUIDS), userUUIDS)

	fmt.Println("查询时间: ", time.Now().Format("2006-01-02 15:04:05"))
	// 创建CSV文件，用于保存记录
	csvFile, err := os.Create("tcms_statistics_" + strconv.Itoa(int(time.Now().Unix())) + ".csv") //创建文件
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile.Close()
	_, _ = csvFile.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	w := csv.NewWriter(csvFile)                //创建一个新的写入文件流
	data := []string{
		"病例号", // 病人ID
		"35<T≤36℃时长",
		"T>37.5℃时长",
		"37.5<T≤38.0℃时长",
		"38.0<T≤38.5℃时长",
		"T>38.5℃时长",
		"最高体温",
		"术后总测量时长",
		"是否寒战",
		"是否谵妄",
	}
	err = w.Write(data)
	if err != nil {
		checkErr(err)
	}
	w.Flush()

	for k, v := range userUUIDS {
		//stmt, err := db.Prepare("SELECT * FROM " + TEMP_TABLE_NAME + "WHERE hardware_sn=?" + " ORDER BY time ASC")
		rows, err := db.Query("SELECT * FROM " + TEMP_TABLE_NAME + " WHERE user_uuid =" + "\"" + k + "\"" + " ORDER BY time ASC")
		//rows, err := stmt.Exec(k)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//column, err := rows.Columns()
		checkErr(err)
		//for _, name := range column {
		//	fmt.Print(name, " ")
		//}
		//fmt.Println()

		fmt.Println("查询开始时间: ", time.Now().Format("2006-01-02 15:04:05"))
		var tcmsData raiingTCMSTempData
		var tempDataArray []temperatureData
		var count int
		var continueTime int64     // 持续时间
		var between350And360 int64 // 大于35，小于等于36
		var exceed375Time int64    // 大于37.5
		var between375And380 int64 // 大于37.5，小于等于38
		var between380And385 int64 // 大于38，小于等于38.5
		var exceed385Time int64    // 大于38.5
		var maxTemperature int     // 最高温度

		startTime := time.Now().Unix() // 温度开始时间，
		endTime := int64(0)            // 温度结束时间

		for rows.Next() {
			err := rows.Scan(&tcmsData.id, &tcmsData.uuid, &tcmsData.userUuid, &tcmsData.time, &tcmsData.timeZone, &tcmsData.hardwareSn, &tcmsData.hardwareVersion,
				&tcmsData.firmwareVersion, &tcmsData.b2wSn, &tcmsData.b2wVersion, &tcmsData.algorithmVersion, &tcmsData.appVersion, &tcmsData.dataVersion,
				&tcmsData.tempMin, &tcmsData.tempMax, &tcmsData.tempAvg, &tcmsData.tempValid, &tcmsData.addTime, &tcmsData.lastUpdateTime, &tcmsData.data)
			checkErr(err)
			if startTime > tcmsData.time {
				startTime = tcmsData.time
			}
			if endTime < tcmsData.time {
				endTime = tcmsData.time
			}
			//fmt.Printf("%s is %s\n", tcmsData.hardwareSn, tcmsData.b2wSn)
			//fmt.Println(tcmsData.b2wSn, tcmsData.data)
			//s := `{"time":1540859820,"temp":36842,"wear_quality":81,"stable":3,"wear_stage":1,"add_time":1540859881}`
			err = json.Unmarshal([]byte(tcmsData.data), &tempDataArray)
			for _, data := range tempDataArray {
				temperatureValue := data.Temp
				if temperatureValue > 35000 && temperatureValue <= 36000 {
					between350And360 += TEMPERATURE_INTERVAL
				}
				if temperatureValue > 37500 {
					exceed375Time += TEMPERATURE_INTERVAL
				}
				if temperatureValue > 37500 && temperatureValue <= 38000 {
					between375And380 += TEMPERATURE_INTERVAL
				}
				if temperatureValue > 38000 && temperatureValue <= 38500 {
					between380And385 += TEMPERATURE_INTERVAL
				}
				if temperatureValue > 38500 {
					exceed385Time += TEMPERATURE_INTERVAL
				}
				continueTime += TEMPERATURE_INTERVAL
				if maxTemperature < temperatureValue {
					maxTemperature = temperatureValue
				}
			}
			//fmt.Println(tempDataArray)
			checkErr(err)
			count ++
		}
		fmt.Println("查询结束时间: ", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println("数据库记录条数: ", count)
		//// 总跨度时间
		//fmt.Println(time.Unix(startTime, 0).Format("2006-01-02 15:04:05"), time.Unix(endTime, 0).Format("2006-01-02 15:04:05"), endTime-startTime)
		//fmt.Println("35<T≤36℃时长: ", between350And360)
		//fmt.Println("T>37.5℃时长: ", exceed375Time)
		//fmt.Println("37.5<T≤38.0℃时长: ", between375And380)
		//fmt.Println("38.0<T≤38.5℃时长: ", between380And385)
		//fmt.Println("T>38.5℃时长: ", exceed385Time)
		//fmt.Println("最高体温: ", maxTemperature)
		//fmt.Println("术后总测量时长: ", continueTime)
		//fmt.Println(startTime, endTime, endTime-startTime)

		//if err := rows.Err(); err != nil {
		//	log.Fatal(err)
		//}

		if err != nil {
			panic(err)
		}
		// 查询事件
		rows2, err := db.Query("SELECT * FROM " + EVENT_TABLE_NAME + " WHERE cases_id =" + "\"" + v + "\"")
		if err != nil {
			log.Fatal(err)
		}
		var event raiingTCMSEventData
		var hangzhanCount int //寒战
		var zhanwangCount int //谵妄
		for rows2.Next() {
			err := rows2.Scan(&event.id, &event.uuid, &event.casesId, &event.userUuid,
				&event.eventUuid, &event.eventType, &event.timezone, &event.startTime,
				&event.endTime, &event.createTime, &event.updateTime, &event.detail,
				&event.addTime, &event.lastUpdateTime)
			checkErr(err)
			if event.eventType == 1018 || event.eventType == 3006 { // 出现寒战事件
				hangzhanCount ++
			} else if event.eventType == 1019 || event.eventType == 3009 { // 出现谵妄事件
				zhanwangCount++
			}
		}

		w := csv.NewWriter(csvFile) //创建一个新的写入文件流
		data := []string{
			v, // 病人ID
			strconv.Itoa(int(between350And360)),
			strconv.Itoa(int(exceed375Time)),
			strconv.Itoa(int(between375And380)),
			strconv.Itoa(int(between380And385)),
			strconv.Itoa(int(exceed385Time)),
			strconv.Itoa(int(maxTemperature)),
			strconv.Itoa(int(continueTime)),
			strconv.Itoa(int(hangzhanCount)),
			strconv.Itoa(int(zhanwangCount)),
		}
		err = w.Write(data)
		if err != nil {
			checkErr(err)
		}
		w.Flush()

	}
	fmt.Println("结束时间: ", time.Now().Format("2006-01-02 15:04:05"))

}
