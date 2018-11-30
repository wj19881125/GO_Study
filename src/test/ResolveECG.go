package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var (
	// 病历号
	mCaseNumber string
	// 开始时间
	mStartTime int64
	// 结束时间
	mEndTime int64
	// 输出的路径
	mOutputPath string

	// 数据库
	mDB *sql.DB
	// 文件
	mOutputRawDataFile *os.File
	mOutputHrFile      *os.File

	// 上一个输出数据的时间 单位 s
	mLastOutputRawDataTime int
	// 下一个输出数据的序号
	mNextOutputRawDataIndex int
)

func main() {
	flag.StringVar(&mCaseNumber, "caseNumber", "", "要查询的病历号")
	flag.Int64Var(&mStartTime, "startTime", 0, "查询数据的开始时间")
	flag.Int64Var(&mEndTime, "endTime", 1, "查询数据的结束时间")
	flag.StringVar(&mOutputPath, "outputPath", "", "存储结果的路径")
	flag.Parse()

	//// 测试用
	//mCaseNumber = "2018111501"
	//mStartTime = 1542955980
	//mEndTime = 1542956160
	//mOutputPath = "/Users/xiaoxin/Desktop"

	fmt.Println("--->病历号:", mCaseNumber)
	fmt.Println("--->开始时间:", mStartTime)
	fmt.Println("--->结束时间:", mEndTime)
	fmt.Println("--->输出路径:", mOutputPath)
	if mCaseNumber == "" {
		fmt.Println("--->失败，未指定病历号！")
		return
	}
	if mOutputPath == "" {
		fmt.Println("--->失败，未指定输出路径！")
		return
	}

	// 打开数据库
	connectDB()
	// 查询User
	userUUID := queryUser()
	if userUUID != "" {
		// 查询数据
		queryEcgRawData(userUUID)
		queryEcgData(userUUID)
	} else {
		fmt.Println("--->失败，未找到指定的病人！")
	}
	// 关闭数据库
	closeDB()
	// 关闭输出文件
	closeOutputFile()
}

// 打开数据库
func connectDB() {
	// 本地
	//db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/raiing_mpcms_v6")
	// MPCMS服务器
	db, err := sql.Open("mysql", "wanglele:f8782ae2@tcp(192.168.1.209:3306)/raiing_mpcms_v6")

	if err != nil {
		fmt.Println("--->connectDB:尝试打开数据库失败：", err.Error())
	} else {
		fmt.Println("--->connectDB:打开数据库成功")
		// 保存
		mDB = db
	}
}

// 关闭数据库
func closeDB() {
	if mDB != nil {
		err := mDB.Close()
		if err != nil {
			fmt.Println("--->closeDB:关闭数据库失败：", err.Error())
		} else {
			fmt.Println("--->closeDB:关闭数据库成功")
		}
		mDB = nil
	}
}

// 查询User
func queryUser() string {
	if mDB == nil {
		fmt.Println("--->queryUser:mDB为nil")
		return ""
	}
	// 查询
	rows, err := mDB.Query("select uuid, bed_num, name from raiing_mpcms_user WHERE case_num = ?;", mCaseNumber)
	if err != nil {
		log.Fatal(err)
		fmt.Println("--->queryUser: 查询时出错：", err.Error())
		return ""
	}
	// 解析行
	userUUID := ""
	for rows.Next() {
		uuid, bedNum, name := "", -1, ""
		err := rows.Scan(&uuid, &bedNum, &name)
		if err != nil {
			fmt.Println("--->queryUser: 解析行出错：", err.Error())
		}
		fmt.Printf("--->查找到对应的病人，病历号：%v，UUID：%v，床号：%d，姓名：%v\n", mCaseNumber, uuid, bedNum, name)
		userUUID = uuid
	}
	defer rows.Close()

	return userUUID
}

// 查询ECG原始数据
func queryEcgRawData(userUUID string) {
	if mDB == nil {
		fmt.Println("--->queryEcgRawData:mDB为nil")
		return
	}
	// 查询
	rows, err := mDB.Query("select time, data_aa, data_ab from raiing_mpcms_ecg_raw_data WHERE user_uuid = ? AND time >= ? AND time <= ?;", userUUID, mStartTime, mEndTime)
	if err != nil {
		log.Fatal(err)
		fmt.Println("--->queryEcgRawData: 查询时出错：", err.Error())
		return
	}
	// 解析行
	for rows.Next() {
		time := 0
		var dataAa []byte
		var dataAb []byte
		err := rows.Scan(&time, &dataAa, &dataAb)
		if err != nil {
			fmt.Println("--->queryEcgRawData: 解析行出错：", err.Error())
		}
		fmt.Println("--->查找到原始ECG数据：", time)
		// 解析数据
		resolveEcgRawData(time, dataAa, dataAb)
	}
	defer rows.Close()
}

// 解析ECG原始数据
func resolveEcgRawData(time int, dataAa []byte, dataAb []byte) {
	if dataAa != nil || dataAb != nil {
		var ecgValueList []int
		// 先解析上半截
		partA, err := base64.StdEncoding.DecodeString(string(dataAa))
		if err == nil {
			partACount := len(partA) / 3
			for i := 0; i < partACount; i++ {
				ecgValue := getInteger(partA[i*3], partA[i*3+1], partA[i*3+2])
				ecgValueList = append(ecgValueList, ecgValue)
			}
		} else {
			fmt.Println("--->queryEcgRawData:解析上半部分的Base64失败：", err)
		}

		// 再解析下半截
		partB, err := base64.StdEncoding.DecodeString(string(dataAb))
		if err == nil {
			partBCount := len(partB) / 3
			for i := 0; i < partBCount; i++ {
				ecgValue := getInteger(partB[i*3], partB[i*3+1], partB[i*3+2])
				ecgValueList = append(ecgValueList, ecgValue)
			}
		} else {
			fmt.Println("--->queryEcgRawData:解析下半部分的Base64失败：", err)
		}
		//fmt.Println("--->queryEcgRawData:ECG数值数量：", len(partA), len(partB))
		fmt.Println("--->当前分钟，ECG的数量：", len(ecgValueList))
		// 输出
		outputRawEcg(time, ecgValueList)
	}
}

// 输出Raw ECG
func outputRawEcg(time int, ecgList []int) {
	outputFile := getOutputRawEcgFile()
	if outputFile == nil {
		fmt.Println("--->outputRawEcg:获取File失败")
		return
	}

	// 调整序号
	if (time - mLastOutputRawDataTime) != 60 {
		mNextOutputRawDataIndex = 0
		fmt.Println("--->时间出现了跳跃，需要将序号重置为0")
	}
	mLastOutputRawDataTime = time

	// 构建输出文本
	contentStr := ""
	for _, item := range ecgList {
		contentStr += fmt.Sprintf("%d,%d,500\n", item, mNextOutputRawDataIndex)
		// 增加序号
		mNextOutputRawDataIndex ++
		if mNextOutputRawDataIndex >= 500 {
			mNextOutputRawDataIndex = 0
		}
	}
	// 输出
	_, err := outputFile.WriteString(contentStr)
	if err != nil {
		fmt.Println("--->outputRawEcg:写File失败")
	}
}

// 获取输出RawECG的File
func getOutputRawEcgFile() *os.File {
	if mOutputRawDataFile == nil {
		// 创建文件
		outputFilePath := mOutputPath + "/RawEcgResult.csv"
		// 删除已经存在的文件
		_, err := os.Stat(outputFilePath)
		if err == nil || os.IsExist(err) {
			fmt.Println("--->删除已经存在的结果文件")
			err := os.Remove(outputFilePath)
			if err != nil {
				fmt.Println("--->outputRawEcg: 删除已经存在的结果文件失败：" + err.Error())
				return nil
			}
		}
		// 新建结果文件
		file, err := os.Create(outputFilePath)
		if err != nil {
			fmt.Println("--->outputCacheTemp: 新建结果文件失败：" + err.Error())
			return nil
		}
		mOutputRawDataFile = file
	}
	return mOutputRawDataFile
}

// 关闭输出文件
func closeOutputFile() {
	if mOutputRawDataFile != nil {
		err := mOutputRawDataFile.Close()
		if err != nil {
			fmt.Println("--->关闭文件（mOutputRawDataFile）出错：", err)
		}
	}
	if mOutputHrFile != nil {
		err := mOutputHrFile.Close()
		if err != nil {
			fmt.Println("--->关闭文件（mOutputHrFile）出错：", err)
		}
	}
}

// 把3byte转化为对应的数值
func getInteger(byte1 byte, byte2 byte, byte3 byte) int {
	number := int(byte1&0Xff) | (int(byte2&0Xff) << 8) | (int(byte3&0Xff) << 16)
	if (number & 0X800000) != 0 {
		number = -(0xFFFFFF - number + 1)
	}
	return number
}

// 查询ECG数据
func queryEcgData(userUUID string) {
	if mDB == nil {
		fmt.Println("--->queryEcgData:mDB为nil")
		return
	}
	// 查询
	rows, err := mDB.Query("select time, heart_rate_data from raiing_mpcms_ecg_data WHERE user_uuid = ? AND time >= ? AND time <= ?;", userUUID, mStartTime, mEndTime)
	if err != nil {
		log.Fatal(err)
		fmt.Println("--->queryEcgData:查询时出错：", err.Error())
		return
	}
	// 解析行
	for rows.Next() {
		time := 0
		var hrJson string
		err := rows.Scan(&time, &hrJson)
		if err != nil {
			fmt.Println("--->queryEcgData:解析行出错：", err.Error())
		}
		fmt.Println("--->查找到ECG数据：", time)
		// 解析数据
		resolveEcgData(time,hrJson)
	}
	defer rows.Close()
}

// 解析ECG数据
func resolveEcgData(time int, hrJson string) {
	if hrJson != "" {
		var hrList []int
		err := json.Unmarshal([]byte(hrJson), &hrList)
		if err == nil {
			fmt.Println("--->当前分钟，HR的数量：", len(hrList))
			// 输出ECG
			outputEcg(time, hrList)
		} else {
			fmt.Println("--->解析Json失败：", err)
		}
	}
}

// 输出ECG
func outputEcg(time int, hrList []int) {
	outputHrFile := getOutputHRFile()
	if outputHrFile == nil {
		fmt.Println("--->outputEcg:获取File失败")
		return
	}

	// 构建输出文本-HR
	hrContentStr := ""
	for index, item := range hrList {
		hrContentStr += fmt.Sprintf("%d,%d\n", time + index, item)
	}
	// 输出
	_, err := outputHrFile.WriteString(hrContentStr)
	if err != nil {
		fmt.Println("--->outputEcg:写File失败")
	}
}

// 获取输出HR的File
func getOutputHRFile() *os.File {
	if mOutputHrFile == nil {
		// 创建文件
		outputFilePath := mOutputPath + "/HrResult.csv"
		// 删除已经存在的文件
		_, err := os.Stat(outputFilePath)
		if err == nil || os.IsExist(err) {
			fmt.Println("--->删除已经存在的结果文件")
			err := os.Remove(outputFilePath)
			if err != nil {
				fmt.Println("--->getOutputHRFile:删除已经存在的结果文件失败：" + err.Error())
				return nil
			}
		}
		// 新建结果文件
		file, err := os.Create(outputFilePath)
		if err != nil {
			fmt.Println("--->getOutputHRFile:新建结果文件失败：" + err.Error())
			return nil
		}
		mOutputHrFile = file
	}
	return mOutputHrFile
}
