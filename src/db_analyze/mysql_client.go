package db_analyze

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
type PostSt struct {
	sex              bool // 男true，女false
	sexInt           int  // 男true，女false
	below360         bool
	between375And380 bool
	between380And385 bool
	exceed385        bool
	isHanzhan        bool
	isZhanwang       bool
	inHospitalTime   int64 // 进入病房时间
}

func AnalyzePostOperationData(user, password, ip, dbName string, ch chan<- StData) {
	if user == "" || password == "" || ip == "" || dbName == "" {
		fmt.Println("传入的用户名等信息为空")
		return
	}
	dbw := DbWorker{
		//Dsn: "root:123456@tcp(127.0.0.1:3306)/raiing_tcms_v6_temp",
		Dsn: user + ":" + password + "@tcp(" + ip + ")/" + dbName,
	}
	db, err := sql.Open("mysql",
		dbw.Dsn)
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("数据库打开成功！")
	defer func() {
		err = db.Close()
		checkErr(err)
	}()
	rows3, err := db.Query("SELECT * FROM " + UserTableName)
	if err != nil {
		log.Fatal(err)
	}
	var userData raiingTCMSUser
	userUUIDS := make(map[string]string, 10) // 用户UUID
	postST := make(map[string]*PostSt, 10)

	for rows3.Next() {
		err := rows3.Scan(&userData.id, &userData.uuid, &userData.caseNum, &userData.bedNum, &userData.name,
			&userData.sex, &userData.birthday, &userData.model, &userData.height, &userData.weight,
			&userData.inHospitalTime, &userData.pacing, &userData.hospital, &userData.department, &userData.status,
			&userData.addTime, &userData.addId, &userData.lastUpdateTime, &userData.lastUpdateId)
		checkErr(err)
		userUUIDS[userData.uuid] = userData.caseNum
		if userData.sex == 1 { // 1为男性
			postST[userData.uuid] = &PostSt{sex: true, sexInt: userData.sex, inHospitalTime: userData.inHospitalTime}
		} else { // 2为女性
			postST[userData.uuid] = &PostSt{sex: false, sexInt: userData.sex, inHospitalTime: userData.inHospitalTime}
		}
	}
	fmt.Println("用户UUID: ", len(userUUIDS), userUUIDS)
	fmt.Println("查询时间: ", time.Now().Format("2006-01-02 15:04:05"))

	//userTempST := make(chan UserTempDistribution, 20)
	// 异步把温度分布写入文件
	//go func() {
	//	saveUserTempDistribution(userTempST)
	//}()
	userUUIDCount := 0
	for k, v := range userUUIDS {
		//stmt, err := db.Prepare("SELECT * FROM " + TEMP_TABLE_NAME + "WHERE hardware_sn=?" + " ORDER BY time ASC")
		// 只查询出病房的数据
		rows, err := db.Query("SELECT * FROM " + TempTableName + " WHERE user_uuid =" + "\"" + k + "\"" + "AND time >" +
			strconv.Itoa(int(postST[k].inHospitalTime)) + " ORDER BY time ASC")
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
				// 只统计稳定以后的温度
				if data.Stable < 2 {
					continue
				}
				temperatureValue := data.Temp
				if temperatureValue < 36000 {
					below360 += TemperatureInterval
				}
				if temperatureValue > 35000 && temperatureValue <= 36000 {
					between350And360 += TemperatureInterval
				}
				if temperatureValue > 37500 {
					exceed375Time += TemperatureInterval
				}
				if temperatureValue > 37500 && temperatureValue <= 38000 {
					between375And380 += TemperatureInterval
				}
				if temperatureValue > 38000 && temperatureValue <= 38500 {
					between380And385 += TemperatureInterval
				}
				if temperatureValue > 38500 {
					exceed385Time += TemperatureInterval
				}
				continueTime += TemperatureInterval
				if maxTemperature < temperatureValue {
					maxTemperature = temperatureValue
				}
			}
			//fmt.Println(tempDataArray)
			checkErr(err)
			count ++
		}
		fmt.Println("查询体温数据结束时间: ", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println("数据库记录条数: ", count)
		if err != nil {
			panic(err)
		}
		// 查询事件
		rows2, err := db.Query("SELECT * FROM " + EventTableName + " WHERE cases_id =" + "\"" + v + "\"")
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

		//userTempDistribution := UserTempDistribution{v,
		//	between350And360,
		//	below360,
		//	exceed375Time,
		//	between375And380,
		//	between380And385,
		//	exceed385Time,
		//	maxTemperature,
		//	continueTime,
		//	hangzhanCount,
		//	zhanwangCount}
		//userTempST <- userTempDistribution
		var post = postST[k]
		stData := StData{caseID: v,
			sex:                           post.sexInt,
			between350And360PostOperation: between350And360,
			below360PostOperation:         below360,
			exceed375TimePostOperation:    exceed375Time,
			between375And380PostOperation: between375And380,
			between380And385PostOperation: between380And385,
			exceed385TimePostOperation:    exceed385Time,
			maxTemperaturePostOperation:   maxTemperature,
			continueTimePostOperation:     continueTime,
			hangzhanCountPostOperation:    hangzhanCount,
			zhanwangCountPostOperation:    zhanwangCount,
		}
		// 输出数据
		ch <- stData

		//var post = postST[k]
		//if below360 > 0 {
		//	post.below360 = true
		//}
		//if between375And380 > 0 {
		//	post.between375And380 = true
		//}
		//if between380And385 > 0 {
		//	post.between380And385 = true
		//}
		//if exceed385Time > 0 {
		//	post.exceed385 = true
		//}
		//if hangzhanCount > 0 {
		//	post.isHanzhan = true
		//}
		//if zhanwangCount > 0 {
		//	post.isZhanwang = true
		//}
		userUUIDCount ++
		//if userUUIDCount > 10 {
		//	break
		//}
		fmt.Println("用户个数: ", userUUIDCount, " ,查询结束时间: ", time.Now().Format("2006-01-02 15:04:05"))
	}
	fmt.Println("结束时间: ", time.Now().Format("2006-01-02 15:04:05"))
	// 关闭channel
	//close(userTempST)
	// 产生统计表格
	//integratedAnalyze(postST)
	close(ch) // 关闭通道
}

type UserTempDistribution struct {
	caseID           string
	between350And360 int64
	below360         int64
	exceed375Time    int64
	between375And380 int64
	between380And385 int64
	exceed385Time    int64
	maxTemperature   int
	continueTime     int64
	hangzhanCount    int
	zhanwangCount    int
}

// 保存用户温度分布
func saveUserTempDistribution(ch chan UserTempDistribution) {
	// 创建CSV文件，用于保存记录。术后的体温监测分布统计
	csvFile, err := os.Create("手术后温度分布统计_" + GetTimeNow() + ".csv") //创建文件
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
	for {
		data, ok := <-ch
		if ok {
			w := csv.NewWriter(csvFile) //创建一个新的写入文件流
			dataString := []string{
				data.caseID, // 病人ID
				fmt.Sprintf("%.2f", float64(data.between350And360)/60),
				fmt.Sprintf("%.2f", float64(data.exceed375Time)/60),
				fmt.Sprintf("%.2f", float64(data.between375And380)/60),
				fmt.Sprintf("%.2f", float64(data.between380And385)/60),
				fmt.Sprintf("%.2f", float64(data.exceed385Time)/60),
				strconv.Itoa(int(data.maxTemperature)),
				fmt.Sprintf("%.2f", float64(data.continueTime)/3600), // 小时
				strconv.Itoa(int(data.hangzhanCount)),
				strconv.Itoa(int(data.zhanwangCount)),
			}
			err = w.Write(dataString)
			if err != nil {
				checkErr(err)
			}
			w.Flush()
		} else {
			fmt.Println("从channel读取温度数据失败")
		}
	}
}

// 整体分析
func integratedAnalyze(postST map[string]*PostSt) {
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
	// 创建CSV文件，用于保存记录, 术后的概率统计
	csvFile1, err := os.Create("手术后温度概率统计_" + GetTimeNow() + ".csv") //创建文件
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
			fmt.Sprintf("%.2f", float64(below360MaleCount)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(between375And380MaleCount)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(between380And385MaleCount)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(exceed385MaleCount)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(hanzhanMaleCount)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(zhanwangMaleCount)/float64(maleCount)),
		},
		{
			"女",
			fmt.Sprintf("%.2f", float64(below360FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between375And380FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between380And385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(exceed385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(hanzhanFemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(zhanwangFemaleCount)/float64(femaleCount)),
		},
		{
			"全部",
			fmt.Sprintf("%.2f", float64(below360MaleCount)/float64(maleCount)+float64(below360FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between375And380MaleCount)/float64(maleCount)+float64(between375And380FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between380And385MaleCount)/float64(maleCount)+float64(between380And385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(exceed385MaleCount)/float64(maleCount)+float64(exceed385FemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(hanzhanMaleCount)/float64(maleCount)+float64(hanzhanFemaleCount)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(zhanwangMaleCount)/float64(maleCount)+float64(zhanwangFemaleCount)/float64(femaleCount)),
		},
	}
	err = w1.WriteAll(data1)
	if err != nil {
		checkErr(err)
	}
	w1.Flush()
}
