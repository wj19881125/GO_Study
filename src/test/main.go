package main

import (
	"db_analyze_pro"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

//var a = 12
// const A = 12
//var b = "string"

//var c interface{}
//
//type d interface {
//
//}
//type e int

var arr [10] int

var slice [6]int

var m map[int]string

type KK struct {
	a int
	b int
}

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

func main() {
	fmt.Println(db_analyze_pro.GetTimeNow())
}

func main01() {
	csvFile, err := os.Open("X:/Golang/GO_Study/trunk/temp_file.csv") //创建文件
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
	for k := range stDatas {
		fmt.Println(k)
	}

	//data := make(chan []string, 10)
	//
	//count := 0
	//
	//go func() {
	//	for {
	//		count++
	//		data <- []string{"raiing" + strconv.Itoa(count)}
	//		fmt.Println("子协程", "raiing"+strconv.Itoa(count))
	//	}
	//}()
	//
	//for {
	//	data2, ok := <-data
	//	if ok {
	//		fmt.Println("主协程，", data2)
	//	}
	//	time.Sleep(time.Second)
	//}

	//return
	//inta := 10
	//intb := 3
	////fmt.Println(float64(inta) / float64(intb))
	//s := fmt.Sprintf("%f", float64(inta)/float64(intb))
	//fmt.Println(s)
	//
	//cc := KK{a: 12}
	//fmt.Println(cc)
	//cc.b = 13
	//fmt.Println(cc)
	//dd := make(map[string]*KK, 10)
	//dd["1223"] = &KK{a: 12}
	//fmt.Println(dd)
	//ee := dd["1223"]
	//ee.b = 14
	//for k, v := range dd {
	//	fmt.Println(k, v.a, v.b)
	//}
	//
	//return
	//m = make(map[int]string)
	//m[1] = "1"
	//m[2] = "2"
	//for index := range m {
	//	fmt.Println(index, m[index])
	//}
	//delete(m, 2)
	//data, ok := m[2]
	//if ok {
	//	fmt.Println(ok, data)
	//} else {
	//	fmt.Println(ok)
	//}

	//for i := 0; i < len(arr); i++ {
	//	fmt.Println(arr[i])
	//}
	//for i, data := range arr {
	//	fmt.Printf("索引: %d, 值: %d", i, data)
	//	fmt.Printf("\n")
	//}
	//slice = make([]int, 10)
	//slice[8] = 10
	//slice = [6]int{1, 2}
	//numbers := append(slice, 11)
	//fmt.Println("切片长度: " + strconv.Itoa(len(slice)) + "切片容量: " + strconv.Itoa(cap(slice)))
	//fmt.Println("切片长度: " + strconv.Itoa(len(numbers)) + "切片容量: " + strconv.Itoa(cap(numbers)))
	//ch := make(chan int)
	//count := 0
	//go func() {
	//	for {
	//		fmt.Println(<-ch)
	//	}
	//}()
	//
	//for {
	//	count ++
	//	ch <- count
	//	time.Sleep(time.Second * 2)
	//}

	//fmt.Println("hello world!")
	//fmt.Println("你好",a,b)
	//fmt.Println("你好",c)
	//fmt.Println("你好",new(d))
	//e := 100
	//fmt.Println("你好",e)

}
