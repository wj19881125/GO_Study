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
	name           string
	sex            int
	birthday       interface{}
	model          int
	height         int
	weight         int
	inHospitalTime int64
	pacing         int
	hospital       interface{}
	department     interface{}
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

// 术后统计
type POST_ST struct {
	sex              bool // 男true，女false
	below360         bool
	between375And380 bool
	between380And385 bool
	exceed385        bool
	isHanzhan        bool
	isZhanwang       bool
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

	rows3, err := db.Query("SELECT * FROM " + USER_TABLE_NAME)
	if err != nil {
		log.Fatal(err)
	}
	var userData raiingTCMSUser
	userUUIDS := make(map[string]string, 10) // 用户UUID
	postST := make(map[string]POST_ST, 10)

	for rows3.Next() {
		err := rows3.Scan(&userData.id, &userData.uuid, &userData.caseNum, &userData.bedNum, &userData.name,
			&userData.sex, &userData.birthday, &userData.model, &userData.height, &userData.weight,
			&userData.inHospitalTime, &userData.pacing, &userData.hospital, &userData.department, &userData.status,
			&userData.addTime, &userData.addId, &userData.lastUpdateTime, &userData.lastUpdateId)
		checkErr(err)
		userUUIDS[userData.uuid] = userData.caseNum
		if userData.sex == 1 { // 1为男性
			postST[userData.uuid] = POST_ST{sex: true}
		} else { // 2为女性
			postST[userData.uuid] = POST_ST{sex: false}
		}
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
		var below360 int64         // 低于36
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
			err = json.Unmarshal([]byte(tcmsData.data), &tempDataArray)
			for _, data := range tempDataArray {
				temperatureValue := data.Temp
				if temperatureValue < 36000 {
					below360 += TEMPERATURE_INTERVAL
				}
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

		var post = postST[k]
		if below360 > 0 {
			post.below360 = true
		}
		if between375And380 > 0 {
			post.between375And380 = true
		}
		if between380And385 > 0 {
			post.between380And385 = true
		}
		if exceed385Time > 0 {
			post.exceed385 = true
		}
		if hangzhanCount > 0 {
			post.isHanzhan = true
		}
		if zhanwangCount > 0 {
			post.isZhanwang = true
		}
	}
	fmt.Println("结束时间: ", time.Now().Format("2006-01-02 15:04:05"))
	maleCount := 0   // 男患者个数
	femaleCount := 0 // 女患者个数
	below360MaleCount := 0
	between375And380MaleCount := 0
	between380And385MaleCount := 0
	exceed385MaleCount := 0
	hanzhanMaleCount := 0
	zhanwangMaleCount := 0
	below360FemaleCount := 0
	between375And380FemaleCount := 0
	between380And385FemaleCount := 0
	exceed385FemaleCount := 0
	hanzhanFemaleCount := 0
	zhanwangFemaleCount := 0
	for _, v := range postST {
		if v.sex {
			maleCount++
			if v.below360 {
				below360MaleCount++
			}
			if v.between375And380 {
				between375And380MaleCount++
			}
			if v.between380And385 {
				between380And385MaleCount++
			}
			if v.exceed385 {
				exceed385MaleCount++
			}
			if v.isHanzhan {
				hanzhanMaleCount++
			}
			if v.isZhanwang {
				zhanwangMaleCount++
			}
		} else {
			femaleCount++
			if v.below360 {
				below360FemaleCount++
			}
			if v.between375And380 {
				between375And380FemaleCount++
			}
			if v.between380And385 {
				between380And385FemaleCount++
			}
			if v.exceed385 {
				exceed385FemaleCount++
			}
			if v.isHanzhan {
				hanzhanFemaleCount++
			}
			if v.isZhanwang {
				zhanwangFemaleCount++
			}
		}
	}
	fmt.Println("男个数: ", maleCount,
		"低于360个数: ", below360MaleCount,
		"大于375小于380个数: ", between375And380MaleCount,
		"大于380小于385个数: ", between380And385MaleCount,
		"超过385个数: ", exceed385MaleCount,
		"寒战个数: ", hanzhanMaleCount,
		"谵妄个数: ", zhanwangMaleCount)
	fmt.Println("女个数: ", femaleCount,
		"低于360个数: ", below360FemaleCount,
		"大于375小于380个数: ", between375And380FemaleCount,
		"大于380小于385个数: ", between380And385FemaleCount,
		"超过385个数: ", exceed385FemaleCount,
		"寒战个数: ", hanzhanFemaleCount,
		"谵妄个数: ", zhanwangFemaleCount)
	// 创建CSV文件，用于保存记录
	csvFile1, err := os.Create("tcms_statistics1_" + strconv.Itoa(int(time.Now().Unix())) + ".csv") //创建文件
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile1.Close()
	_, _ = csvFile1.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	w1 := csv.NewWriter(csvFile1)               //创建一个新的写入文件流
	data1 := [][]string{
		{"性别",
			"低于36℃发生率",
			"37.5-38.0℃发生率",
			"38.0-38.5℃发生率",
			"高于38.5℃发生率",
			"寒战发生率",
			"谵妄发生率",},
		{
			"男",
			fmt.Sprintf("%f", float64(below360MaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(between375And380MaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(between380And385MaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(exceed385MaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(hanzhanMaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(zhanwangMaleCount)/float64(maleCount)),
		},
		{
			"女",
			fmt.Sprintf("%f", float64(below360FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(between375And380FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(between380And385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(exceed385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(hanzhanFemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(zhanwangFemaleCount)/float64(femaleCount)),
		},
		{
			"全部",
			fmt.Sprintf("%f", float64(below360MaleCount)/float64(maleCount)+float64(below360FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(between375And380MaleCount)/float64(maleCount)+float64(between375And380FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(between380And385MaleCount)/float64(maleCount)+float64(between380And385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(exceed385MaleCount)/float64(maleCount)+float64(exceed385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(hanzhanMaleCount)/float64(maleCount)+float64(hanzhanFemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(zhanwangMaleCount)/float64(maleCount)+float64(zhanwangFemaleCount)/float64(femaleCount)),
		},
	}
	err = w1.WriteAll(data1)
	if err != nil {
		checkErr(err)
	}
	w1.Flush()

}

func gen(postST *map[string]POST_ST) {
	maleCount := 0   // 男患者个数
	femaleCount := 0 // 女患者个数
	below360MaleCount := 0
	between375And380MaleCount := 0
	between380And385MaleCount := 0
	exceed385MaleCount := 0
	hanzhanMaleCount := 0
	zhanwangMaleCount := 0
	below360FemaleCount := 0
	between375And380FemaleCount := 0
	between380And385FemaleCount := 0
	exceed385FemaleCount := 0
	hanzhanFemaleCount := 0
	zhanwangFemaleCount := 0
	for _, v := range postST {
		if v.sex {
			maleCount++
			if v.below360 {
				below360MaleCount++
			}
			if v.between375And380 {
				between375And380MaleCount++
			}
			if v.between380And385 {
				between380And385MaleCount++
			}
			if v.exceed385 {
				exceed385MaleCount++
			}
			if v.isHanzhan {
				hanzhanMaleCount++
			}
			if v.isZhanwang {
				zhanwangMaleCount++
			}
		} else {
			femaleCount++
			if v.below360 {
				below360FemaleCount++
			}
			if v.between375And380 {
				between375And380FemaleCount++
			}
			if v.between380And385 {
				between380And385FemaleCount++
			}
			if v.exceed385 {
				exceed385FemaleCount++
			}
			if v.isHanzhan {
				hanzhanFemaleCount++
			}
			if v.isZhanwang {
				zhanwangFemaleCount++
			}
		}
	}
	fmt.Println("男个数: ", maleCount,
		"低于360个数: ", below360MaleCount,
		"大于375小于380个数: ", between375And380MaleCount,
		"大于380小于385个数: ", between380And385MaleCount,
		"超过385个数: ", exceed385MaleCount,
		"寒战个数: ", hanzhanMaleCount,
		"谵妄个数: ", zhanwangMaleCount)
	fmt.Println("女个数: ", femaleCount,
		"低于360个数: ", below360FemaleCount,
		"大于375小于380个数: ", between375And380FemaleCount,
		"大于380小于385个数: ", between380And385FemaleCount,
		"超过385个数: ", exceed385FemaleCount,
		"寒战个数: ", hanzhanFemaleCount,
		"谵妄个数: ", zhanwangFemaleCount)
	// 创建CSV文件，用于保存记录
	csvFile1, err := os.Create("tcms_statistics1_" + strconv.Itoa(int(time.Now().Unix())) + ".csv") //创建文件
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile1.Close()
	_, _ = csvFile1.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	w1 := csv.NewWriter(csvFile1)               //创建一个新的写入文件流
	data1 := [][]string{
		{"性别",
			"低于36℃发生率",
			"37.5-38.0℃发生率",
			"38.0-38.5℃发生率",
			"高于38.5℃发生率",
			"寒战发生率",
			"谵妄发生率",},
		{
			"男",
			fmt.Sprintf("%f", float64(below360MaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(between375And380MaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(between380And385MaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(exceed385MaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(hanzhanMaleCount)/float64(maleCount)),
			fmt.Sprintf("%f", float64(zhanwangMaleCount)/float64(maleCount)),
		},
		{
			"女",
			fmt.Sprintf("%f", float64(below360FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(between375And380FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(between380And385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(exceed385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(hanzhanFemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(zhanwangFemaleCount)/float64(femaleCount)),
		},
		{
			"全部",
			fmt.Sprintf("%f", float64(below360MaleCount)/float64(maleCount)+float64(below360FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(between375And380MaleCount)/float64(maleCount)+float64(between375And380FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(between380And385MaleCount)/float64(maleCount)+float64(between380And385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(exceed385MaleCount)/float64(maleCount)+float64(exceed385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(hanzhanMaleCount)/float64(maleCount)+float64(hanzhanFemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%f", float64(zhanwangMaleCount)/float64(maleCount)+float64(zhanwangFemaleCount)/float64(femaleCount)),
		},
	}
	err = w1.WriteAll(data1)
	if err != nil {
		checkErr(err)
	}
	w1.Flush()
}
