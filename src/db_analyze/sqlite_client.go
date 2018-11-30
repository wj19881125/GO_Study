package db_analyze

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func checkErr1(err error) {
	if err != nil {
		panic(err)
	}
}

type temperatureData1 struct {
	id               int
	time             int
	timeZone         int
	algVer           string
	temperature      int
	stable           int
	wearScore        int
	wearGrade        int
	source           int
	sn               string
	hardwareVer      string
	firmwareVer      string
	patientId        string
	usageScenario    int
	ctmUsageScenario int
	isUpload         int
}

var filePaths []string
//获取指定目录下的所有文件和目录
func ListDir(dirPth string) (err error) {
	dir, err := ioutil.ReadDir(dirPth)
	PthSep := string(os.PathSeparator)
	// suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			//files1 = append(files1, dirPth+PthSep+fi.Name())
			err := ListDir(dirPth + PthSep + fi.Name())
			checkErr1(err)
			//fmt.Println(dirPth + PthSep + fi.Name())
		} else {
			pathName := dirPth + PthSep + fi.Name()
			//fmt.Println(pathName)
			//files = append(files, dirPth+PthSep+fi.Name())
			if fi.Name() == "TemperatureDbForERAS.db" {
				fmt.Println(pathName)
				filePaths = append(filePaths, pathName)
			}
		}
	}
	return nil
}

func AnalyzeInOperationData() {
	filePaths = make([]string, 0, 10)
	originalPath := "X:/GO/raw_data"
	err := ListDir(originalPath)
	checkErr1(err)
	//
	//dbNames := make([]string, 1, 10)
	//dbNames[0] = "X:/Golang/GO_Study/trunk/TemperatureDbForERAS.db"
	//dbNames = append(dbNames, "TemperatureDbForERAS.db")
	userTempDistributionInOperationChan := make(chan map[string]*UserTempDistributionInOperation, 20)
	// 保存统计结果
	go func() { saveUserTempDistributionInOperation(userTempDistributionInOperationChan) }()
	// 遍历所有数据库
	for _, dbName := range filePaths {
		fmt.Println("数据库名称: ", dbName)
		db, err := sql.Open("sqlite3", dbName)
		//db, err := sql.Open("sqlite3", "./1.db")
		checkErr1(err)
		fmt.Println("打开数据库成功")
		rows, err := db.Query("SELECT * FROM body_temperature ORDER BY time ASC")
		if err != nil {
			checkErr1(err)
			return
		}
		userTempDistributionInOperations := make(map[string]*UserTempDistributionInOperation)
		var data temperatureData1
		for rows.Next() {
			err = rows.Scan(&data.id, &data.time, &data.timeZone, &data.algVer, &data.temperature,
				&data.stable, &data.wearScore, &data.wearGrade, &data.source, &data.sn,
				&data.hardwareVer, &data.firmwareVer, &data.patientId, &data.usageScenario, &data.ctmUsageScenario,
				&data.isUpload)
			if err != nil {
				fmt.Println("查询温度数据出现错误,", err)
				continue
			}
			//strTime := time.Unix(int64(data.time), 0).Format("2006-01-02 15:04:05")
			//strTime := time.Now().Format("2006-01-02 15:04:05")
			//fmt.Print(strTime)
			//fmt.Println(data)
			if data.stable < 2 { // 稳定状态
				continue
			}
			value, ok := userTempDistributionInOperations[data.patientId]
			if ok {
				if value.minTemperature > data.temperature {
					value.minTemperature = data.temperature
				}
				if data.temperature > 35000 && data.temperature <= 36000 {
					value.between350And360 += TemperatureInterval
				}
				if data.temperature > 37500 && data.temperature <= 38000 {
					value.between375And380 += TemperatureInterval
				}
				if data.temperature > 38000 {
					value.exceed38 += TemperatureInterval
				}
				value.continueTime += TemperatureInterval
			} else {
				var userTempDistributionInOperation UserTempDistributionInOperation
				// 最低温度赋初始值
				userTempDistributionInOperation.minTemperature = 70000
				if userTempDistributionInOperation.minTemperature > data.temperature {
					userTempDistributionInOperation.minTemperature = data.temperature
				}
				if data.temperature > 35000 && data.temperature <= 36000 {
					userTempDistributionInOperation.between350And360 += TemperatureInterval
				}
				if data.temperature > 37500 && data.temperature <= 38000 {
					userTempDistributionInOperation.between375And380 += TemperatureInterval
				}
				if data.temperature > 38000 {
					userTempDistributionInOperation.exceed38 += TemperatureInterval
				}
				userTempDistributionInOperation.continueTime += TemperatureInterval
				// 插入元素
				userTempDistributionInOperations[data.patientId] = &userTempDistributionInOperation
			}
		}
		userTempDistributionInOperationChan <- userTempDistributionInOperations
		// 关闭数据库
		err = db.Close()
		checkErr1(err)
	}
	// 保证程序可以正常退出
	time.Sleep(time.Second * 5)
	// 关闭channel
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

func isZip(zipPath string) bool {
	f, err := os.Open(zipPath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 4)
	if n, err := f.Read(buf); err != nil || n < 4 {
		return false
	}

	return bytes.Equal(buf, []byte("PK\x03\x04"))
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}
