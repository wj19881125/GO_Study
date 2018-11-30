package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strconv"
	"time"
)

func checkErr1(err error) {
	if err != nil {
		panic(err)
	}
}

const TEMPERATURE_INTERVAL1 = 4

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

func main() {
	db, err := sql.Open("sqlite3", "TemperatureDbForERAS.db")
	//db, err := sql.Open("sqlite3", "./1.db")
	checkErr1(err)
	defer func() {
		err = db.Close()
		checkErr1(err)
	}()
	fmt.Println("打开数据库成功")
	rows, err := db.Query("SELECT * FROM body_temperature ORDER BY time ASC")
	if err != nil {
		checkErr1(err)
		return
	}
	userTempDistributionInOperationChan := make(chan map[string]*UserTempDistributionInOperation, 20)
	userTempDistributionInOperations := make(map[string]*UserTempDistributionInOperation)
	// 保存统计结果
	go func() { saveUserTempDistributionInOperation(userTempDistributionInOperationChan) }()

	var data temperatureData1
	for rows.Next() {
		err = rows.Scan(&data.id, &data.time, &data.timeZone, &data.algVer, &data.temperature, &data.stable, &data.wearScore,
			&data.wearGrade, &data.source, &data.sn, &data.hardwareVer, &data.firmwareVer, &data.patientId, &data.usageScenario, &data.isUpload)
		checkErr1(err)
		strTime := time.Unix(int64(data.time), 0).Format("2006-01-02 15:04:05")
		//strTime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Print(strTime)
		fmt.Println(data)
		if data.stable < 2 { // 稳定状态
			continue
		}
		value, ok := userTempDistributionInOperations[data.patientId]
		if ok {
			if value.minTemperature > data.temperature {
				value.minTemperature = data.temperature
			}
			if data.temperature > 35000 && data.temperature <= 36000 {
				value.between350And360 += TEMPERATURE_INTERVAL1
			}
			if data.temperature > 37500 && data.temperature <= 38000 {
				value.between375And380 += TEMPERATURE_INTERVAL1
			}
			if data.temperature > 38000 {
				value.exceed38 += TEMPERATURE_INTERVAL1
			}
			value.continueTime += TEMPERATURE_INTERVAL1
		} else {
			var userTempDistributionInOperation UserTempDistributionInOperation
			// 最低温度赋初始值
			userTempDistributionInOperation.minTemperature = 70000
			if userTempDistributionInOperation.minTemperature > data.temperature {
				userTempDistributionInOperation.minTemperature = data.temperature
			}
			if data.temperature > 35000 && data.temperature <= 36000 {
				userTempDistributionInOperation.between350And360 += TEMPERATURE_INTERVAL1
			}
			if data.temperature > 37500 && data.temperature <= 38000 {
				userTempDistributionInOperation.between375And380 += TEMPERATURE_INTERVAL1
			}
			if data.temperature > 38000 {
				userTempDistributionInOperation.exceed38 += TEMPERATURE_INTERVAL1
			}
			userTempDistributionInOperation.continueTime += TEMPERATURE_INTERVAL1
			// 插入元素
			userTempDistributionInOperations[data.patientId] = &userTempDistributionInOperation
		}
	}
	userTempDistributionInOperationChan <- userTempDistributionInOperations
	// 保证程序可以正常退出
	time.Sleep(time.Second * 5)
	close(userTempDistributionInOperationChan)
}

type UserTempDistributionInOperation struct {
	caseID           string // 病例号
	minTemperature   int    // 最低体温
	between350And360 int    // 大于350小于等于360时长
	between375And380 int    // 大于375，小于等于380时长
	exceed38         int    // 超过38度时长
	continueTime     int    // 总测量时长
}

// 保存术中的用户温度分布
func saveUserTempDistributionInOperation(ch chan map[string]*UserTempDistributionInOperation) {
	// 创建CSV文件，用于保存记录。术后的体温监测分布统计
	csvFile, err := os.Create("手术中数据分布_" + strconv.Itoa(int(time.Now().Unix())) + ".csv") //创建文件
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile.Close()
	_, _ = csvFile.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	w := csv.NewWriter(csvFile)                //创建一个新的写入文件流
	data := []string{
		"病例号",  // 病人ID
		"最低体温", // 最低体温
		"35<T≤36℃时长",
		"37.5<T≤38.0℃时长",
		"T>38℃时长",
		"总测量时长",
	}
	err = w.Write(data)
	if err != nil {
		checkErr1(err)
	}
	w.Flush()
	for {
		data, ok := <-ch
		if ok {
			for k, v := range data {
				w := csv.NewWriter(csvFile) //创建一个新的写入文件流
				dataString := []string{
					k, // 病人ID
					strconv.Itoa(int(v.minTemperature)),
					fmt.Sprintf("%.2f", float64(v.between350And360)/60),
					fmt.Sprintf("%.2f", float64(v.between375And380)/60),
					fmt.Sprintf("%.2f", float64(v.exceed38)/60),
					fmt.Sprintf("%.2f", float64(v.continueTime)/3600), // 小时
				}
				err = w.Write(dataString)
				if err != nil {
					checkErr1(err)
				}
				w.Flush()
			}
		} else {
			fmt.Println("从channel读取温度数据失败")
		}
	}
}
