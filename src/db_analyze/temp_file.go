package db_analyze

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

// 统计数据
type StData struct {
	caseID                        string
	sex                           int
	minTemperatureInOperation     int
	between350And360InOperation   int64 // 手术中大于350，低于360
	below360InOperation           int64 // 手术中低于360
	between375And380InOperation   int64 // 手术中大于375，低于380
	exceed380InOperation          int64 // 手术中大于380
	exceed385InOperation          int64 // 手术中大于385
	continueTimeInOperation       int64 // 手术中总测量时长
	between350And360PostOperation int64 // 手术后大于350，小于360时长
	below360PostOperation         int64
	exceed375TimePostOperation    int64
	between375And380PostOperation int64
	between380And385PostOperation int64
	exceed385TimePostOperation    int64
	maxTemperaturePostOperation   int
	continueTimePostOperation     int64
	hangzhanCountPostOperation    int // 手术后寒战次数
	zhanwangCountPostOperation    int // 手术后谵妄次数
}

func IsExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func SaveData(ch <-chan StData) {
	ok, e := IsExists(STDataTempFileName)
	if e != nil {
		fmt.Println("判断文件存在返回失败")
		return
	}
	if ok {
		err := os.Remove(STDataTempFileName)
		if err != nil {
			fmt.Println("删除临时文件失败")
			return
		}
	}
	csvFile, err := os.Create(STDataTempFileName) //创建文件
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile.Close()

	_, _ = csvFile.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	w := csv.NewWriter(csvFile)                //创建一个新的写入文件流
	data := []string{
		"病例号", // 病人ID
		"性别",  // 病人ID
		"术中-最低温度",
		"术中-35.0<T<=36.0℃时长",
		"术中-T<36.0℃时长",
		"术中-37.0<T≤38.0℃时长",
		"术中-38.0<T≤38.5℃时长",
		"术中-T>38.0℃时长",
		"术中-T>38.5℃时长",
		"术中-总测量时长",
		"术后-35.0<T≤36.0℃时长",
		"术后-T≤36.0℃时长",
		"术后-T>37.5℃时长",
		"术后-37.5<T≤38.0℃时长",
		"术后-38.0<T≤38.5℃时长",
		"术后-T>38.5℃时长",
		"术后-最高温度",
		"术后-总测量时长",
		"术后-寒战次数",
		"术后-谵妄次数",
	}
	err = w.Write(data)
	if err != nil {
		checkErr(err)
	}
	w.Flush()

	for {
		if data, ok := <-ch; ok {
			//if data.(type)
			w := csv.NewWriter(csvFile) //创建一个新的写入文件流
			dataString := []string{
				data.caseID, // 病人ID
				strconv.Itoa(int(data.sex)),
				strconv.Itoa(int(data.minTemperatureInOperation)),
				strconv.Itoa(int(data.between350And360InOperation)),
				strconv.Itoa(int(data.below360InOperation)),
				strconv.Itoa(int(data.between375And380InOperation)),
				strconv.Itoa(int(data.exceed380InOperation)),
				strconv.Itoa(int(data.exceed385InOperation)),
				strconv.Itoa(int(data.continueTimeInOperation)),
				strconv.Itoa(int(data.between350And360PostOperation)),
				strconv.Itoa(int(data.below360PostOperation)),
				strconv.Itoa(int(data.exceed375TimePostOperation)),
				strconv.Itoa(int(data.between375And380PostOperation)),
				strconv.Itoa(int(data.between380And385PostOperation)),
				strconv.Itoa(int(data.exceed385TimePostOperation)),
				strconv.Itoa(int(data.maxTemperaturePostOperation)),
				strconv.Itoa(int(data.continueTimePostOperation)),
				strconv.Itoa(int(data.hangzhanCountPostOperation)),
				strconv.Itoa(int(data.zhanwangCountPostOperation)),
			}
			err = w.Write(dataString)
			if err != nil {
				checkErr1(err)
			}
			w.Flush()
		} else {
			fmt.Println("接收不到数据中止循环")
			break
		}

	}
	AnalyzeCSVFile(STDataTempFileName)
}

//type TempDistributionST struct {
//	caseID                      string
//	minTemperatureInOperation   int
//	between350And360InOperation int64 // 手术中大于350，低于360
//	//below360InOperation           int64 // 手术中低于360
//	between375And380InOperation int64 // 手术中大于375，低于380
//	exceed380InOperation        int64 // 手术中大于380
//	//exceed385InOperation          int64 // 手术中大于385
//	continueTimeInOperation       int64 // 手术中总测量时长
//	between350And360PostOperation int64 // 手术后大于350，小于360时长
//	exceed375TimePostOperation    int64
//	between375And380PostOperation int64
//	between380And385PostOperation int64
//	exceed385TimePostOperation    int64
//	maxTemperaturePostOperation   int
//	continueTimePostOperation     int64
//	hangzhanCountPostOperation    int // 手术后寒战次数
//	zhanwangCountPostOperation    int // 手术后谵妄次数
//}

func AnalyzeCSVFile(filePath string) {
	//csvFile, err := os.Open("X:/Golang/GO_Study/trunk/temp_file.csv") //创建文件
	csvFile, err := os.Open(filePath) //创建文件
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile.Close()
	fmt.Println("文件名称," + csvFile.Name())
	//_, _ = csvFile.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM

	stDatas := make(map[string]*StData, 10)
	r := csv.NewReader(csvFile)
	for {
		records, err := r.Read()
		if err == io.EOF {
			fmt.Println(err)
			break
		}
		var stData StData
		//var caseID string
		//for _, data := range records {
		//	print(data, ",")
		//
		//}
		//println()
		stData.caseID = records[0]
		//println(stData.caseID)
		intValue, err := strconv.Atoi(records[1])
		stData.sex = intValue
		intValue, err = strconv.Atoi(records[2])
		stData.minTemperatureInOperation = intValue
		intValue, err = strconv.Atoi(records[3])
		stData.between350And360InOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[4])
		stData.below360InOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[5])
		stData.between375And380InOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[6])
		stData.exceed380InOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[7])
		stData.exceed385InOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[8])
		stData.continueTimeInOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[9])
		stData.between350And360PostOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[10])
		stData.below360PostOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[11])
		stData.exceed375TimePostOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[12])
		stData.between375And380PostOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[13])
		stData.between380And385PostOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[14])
		stData.exceed385TimePostOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[15])
		stData.maxTemperaturePostOperation = intValue
		intValue, err = strconv.Atoi(records[16])
		stData.continueTimePostOperation = int64(intValue)
		intValue, err = strconv.Atoi(records[17])
		stData.hangzhanCountPostOperation = intValue
		intValue, err = strconv.Atoi(records[18])
		stData.zhanwangCountPostOperation = intValue
		value, ok := stDatas[stData.caseID ]
		if ok {
			if value.sex == 0 {
				value.sex = stData.sex
			}
			if value.minTemperatureInOperation == 0 {
				value.sex = stData.minTemperatureInOperation
			}
			if value.maxTemperaturePostOperation == 0 {
				value.maxTemperaturePostOperation = stData.maxTemperaturePostOperation
			}
			value.between350And360InOperation += stData.between350And360InOperation
			value.below360InOperation += stData.below360InOperation
			value.between375And380InOperation += stData.between375And380InOperation
			value.exceed380InOperation += stData.exceed380InOperation
			value.exceed385InOperation += stData.exceed385InOperation
			value.continueTimeInOperation += stData.continueTimeInOperation
			value.between350And360PostOperation += stData.between350And360PostOperation
			value.below360PostOperation += stData.below360PostOperation
			value.exceed375TimePostOperation += stData.exceed375TimePostOperation
			value.between375And380PostOperation += stData.between375And380PostOperation
			value.between380And385PostOperation += stData.between380And385PostOperation
			value.exceed385TimePostOperation += stData.exceed385TimePostOperation
			value.continueTimePostOperation += stData.continueTimePostOperation
			value.hangzhanCountPostOperation += stData.hangzhanCountPostOperation
			value.zhanwangCountPostOperation += stData.zhanwangCountPostOperation
		} else {
			stDatas[stData.caseID] = &stData
		}
	}
	//for k := range stDatas {
	//	fmt.Println(k)
	//}

	SaveUserTempDistribution(stDatas)
	IntegratedAnalyze(stDatas)
}

// 保存用户温度分布
func SaveUserTempDistribution(postST map[string]*StData) {
	// 创建CSV文件，用于保存记录。术后的体温监测分布统计
	csvFile, err := os.Create("温度分布统计_" + GetTimeNow() + ".csv") //创建文件
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile.Close()
	_, _ = csvFile.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	w := csv.NewWriter(csvFile)                //创建一个新的写入文件流
	data := []string{
		"病例号", // 病人ID
		// 术中
		"术中-最低体温",
		"术中-35<T≤36℃时长",
		"术中-37.5<T≤38.0℃时长",
		"术中-T>38℃ 时长",
		"术中-总测量时长",
		// 术后
		"术后-35<T≤36℃时长",
		"术后-T>37.5℃时长",
		"术后-37.5<T≤38.0℃时长",
		"术后-38.0<T≤38.5℃时长",
		"术后-T>38.5℃时长",
		"术后-最高体温",
		"术后-术后总测量时长",
		"术后-是否寒战",
		"术后-是否谵妄",
	}
	err = w.Write(data)
	if err != nil {
		checkErr(err)
	}
	w.Flush()
	for _, data := range postST {
		w := csv.NewWriter(csvFile) //创建一个新的写入文件流
		dataString := []string{
			data.caseID, // 病人ID
			//术中
			strconv.Itoa(int(data.minTemperatureInOperation)),
			fmt.Sprintf("%.2f", float64(data.between350And360InOperation)/60),
			fmt.Sprintf("%.2f", float64(data.between375And380InOperation)/60),
			fmt.Sprintf("%.2f", float64(data.exceed380InOperation)/60),
			fmt.Sprintf("%.2f", float64(data.continueTimeInOperation)/60),
			// 术后
			fmt.Sprintf("%.2f", float64(data.between350And360PostOperation)/60),
			fmt.Sprintf("%.2f", float64(data.exceed375TimePostOperation)/60),
			fmt.Sprintf("%.2f", float64(data.between375And380PostOperation)/60),
			fmt.Sprintf("%.2f", float64(data.between380And385PostOperation)/60),
			fmt.Sprintf("%.2f", float64(data.exceed385TimePostOperation)/60),
			strconv.Itoa(int(data.maxTemperaturePostOperation)),
			fmt.Sprintf("%.2f", float64(data.continueTimePostOperation)/3600), // 小时
			strconv.Itoa(int(data.hangzhanCountPostOperation)),
			strconv.Itoa(int(data.zhanwangCountPostOperation)),
		}
		err = w.Write(dataString)
		if err != nil {
			checkErr(err)
		}
		w.Flush()
	}
}

// 整体分析
func IntegratedAnalyze(postST map[string]*StData) {
	maleCount := 0   // 男患者个数
	femaleCount := 0 // 女患者个数
	// 术中
	below360MaleCountInOperation := 0
	between375And380MaleCountInOperation := 0
	exceed385MaleCountInOperation := 0
	below360FemaleCountInOperation := 0
	between375And380FemaleCountInOperation := 0
	exceed385FemaleCountInOperation := 0
	// 术后
	below360MaleCountPostOperation := 0
	between375And380MaleCountPostOperation := 0
	between380And385MaleCountPostOperation := 0
	exceed385MaleCountPostOperation := 0
	hanzhanMaleCountPostOperation := 0
	zhanwangMaleCountPostOperation := 0
	below360FemaleCountPostOperation := 0
	between375And380FemaleCountPostOperation := 0
	between380And385FemaleCountPostOperation := 0
	exceed385FemaleCountPostOperation := 0
	hanzhanFemaleCountPostOperation := 0
	zhanwangFemaleCountPostOperation := 0
	for _, v := range postST {
		if v.sex == 1 {
			maleCount++
			// 术中
			if v.below360InOperation > 0 {
				below360MaleCountInOperation++
			}
			if v.between375And380InOperation > 0 {
				between375And380MaleCountInOperation++
			}
			if v.exceed385InOperation > 0 {
				exceed385MaleCountInOperation++
			}
			if v.below360PostOperation > 0 {
				below360MaleCountPostOperation++
			}
			if v.between375And380PostOperation > 0 {
				between375And380MaleCountPostOperation++
			}
			if v.between380And385PostOperation > 0 {
				between380And385MaleCountPostOperation++
			}
			if v.exceed385TimePostOperation > 0 {
				exceed385MaleCountPostOperation++
			}
			if v.hangzhanCountPostOperation > 0 {
				hanzhanMaleCountPostOperation++
			}
			if v.zhanwangCountPostOperation > 0 {
				zhanwangMaleCountPostOperation++
			}
		} else if v.sex == 2 {
			femaleCount++
			// 术中
			if v.below360InOperation > 0 {
				below360FemaleCountInOperation++
			}
			if v.between375And380InOperation > 0 {
				between375And380FemaleCountInOperation++
			}
			if v.exceed385InOperation > 0 {
				exceed385FemaleCountInOperation++
			}
			// 术后
			if v.below360PostOperation > 0 {
				below360FemaleCountPostOperation++
			}
			if v.between375And380PostOperation > 0 {
				between375And380FemaleCountPostOperation++
			}
			if v.between380And385PostOperation > 0 {
				between380And385FemaleCountPostOperation++
			}
			if v.exceed385TimePostOperation > 0 {
				exceed385FemaleCountPostOperation++
			}
			if v.hangzhanCountPostOperation > 0 {
				hanzhanFemaleCountPostOperation++
			}
			if v.zhanwangCountPostOperation > 0 {
				zhanwangFemaleCountPostOperation++
			}
		}
	}
	fmt.Println("男个数: ", maleCount,
		"低于360个数: ", below360MaleCountPostOperation,
		"大于375小于380个数: ", between375And380MaleCountPostOperation,
		"大于380小于385个数: ", between380And385MaleCountPostOperation,
		"超过385个数: ", exceed385MaleCountPostOperation,
		"寒战个数: ", hanzhanMaleCountPostOperation,
		"谵妄个数: ", zhanwangMaleCountPostOperation)
	fmt.Println("女个数: ", femaleCount,
		"低于360个数: ", below360FemaleCountPostOperation,
		"大于375小于380个数: ", between375And380FemaleCountPostOperation,
		"大于380小于385个数: ", between380And385FemaleCountPostOperation,
		"超过385个数: ", exceed385FemaleCountPostOperation,
		"寒战个数: ", hanzhanFemaleCountPostOperation,
		"谵妄个数: ", zhanwangFemaleCountPostOperation)
	// 创建CSV文件，用于保存记录, 术后的概率统计
	csvFile1, err := os.Create("温度概率统计_" + GetTimeNow() + ".csv") //创建文件
	if err != nil {
		fmt.Println(err)
		return
	}
	defer csvFile1.Close()
	_, _ = csvFile1.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
	w1 := csv.NewWriter(csvFile1)               //创建一个新的写入文件流
	data1 := [][]string{
		{"性别",
			"术中-低于36℃发生率",
			"术中-37.5-38.0℃发生率",
			"术中-高于38.5℃发生率",
			"术后-低于36℃发生率",
			"术后-37.5-38.0℃发生率",
			"术后-38.0-38.5℃发生率",
			"术后-高于38.5℃发生率",
			"术后-寒战发生率",
			"术后-谵妄发生率",},
		{
			"男",
			fmt.Sprintf("%.2f", float64(below360MaleCountInOperation)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(between375And380MaleCountInOperation)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(exceed385MaleCountInOperation)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(below360MaleCountPostOperation)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(between375And380MaleCountPostOperation)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(between380And385MaleCountPostOperation)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(exceed385MaleCountPostOperation)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(hanzhanMaleCountPostOperation)/float64(maleCount)),
			fmt.Sprintf("%.2f", float64(zhanwangMaleCountPostOperation)/float64(maleCount)),
		},
		{
			"女",
			fmt.Sprintf("%.2f", float64(below360FemaleCountInOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between375And380FemaleCountInOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(exceed385FemaleCountInOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(below360FemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between375And380FemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between380And385FemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(exceed385FemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(hanzhanFemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(zhanwangFemaleCountPostOperation)/float64(femaleCount)),
		},
		{
			"全部",
			fmt.Sprintf("%.2f", float64(below360MaleCountInOperation)/float64(maleCount)+float64(below360FemaleCountInOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between375And380MaleCountInOperation)/float64(maleCount)+float64(between375And380FemaleCountInOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(exceed385MaleCountInOperation)/float64(maleCount)+float64(exceed385FemaleCountInOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(below360MaleCountPostOperation)/float64(maleCount)+float64(below360FemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between375And380MaleCountPostOperation)/float64(maleCount)+float64(between375And380FemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(between380And385MaleCountPostOperation)/float64(maleCount)+float64(between380And385FemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(exceed385MaleCountPostOperation)/float64(maleCount)+float64(exceed385FemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(hanzhanMaleCountPostOperation)/float64(maleCount)+float64(hanzhanFemaleCountPostOperation)/float64(femaleCount)),
			fmt.Sprintf("%.2f", float64(zhanwangMaleCountPostOperation)/float64(maleCount)+float64(zhanwangFemaleCountPostOperation)/float64(femaleCount)),
		},
	}
	err = w1.WriteAll(data1)
	if err != nil {
		checkErr(err)
	}
	w1.Flush()
}
