package main

import (
	"db_analyze"
	"fmt"
	"os"
)

func main() {
	// 解析传入参数
	if len(os.Args) != 5 {
		fmt.Println("参数不够,参数1: 所有手术室数据的根目录；参数2: TCMS系统数据库用户名；参数3: TCMS系统数据库密码；参数4：TCMS数据库服务器IP")
		return
	}
	//fmt.Println(os.Args) //打印切片内容
	//for i := 0; i < len(os.Args); i++ {
	//	fmt.Println(os.Args[i])
	//}
	//return
	// 1. 手术室所有压缩包解压以后的路径
	//rawDataPath := "X:/GO/raw_data"
	rawDataPath := os.Args[1]
	// 2. TCMS服务器IP,数据库用户名和密码
	//user := "root"
	user := os.Args[2]
	//password := "123456"
	password := os.Args[3]
	//ip := "127.0.0.1:3306"
	ip := os.Args[4]
	dbName := "raiing_tcms_v6_20181204"
	// 3. 保存文件名
	// 4. 保存文件路径

	stDatas := make(chan db_analyze.StData, 20)

	go db_analyze.AnalyzeInOperationData(rawDataPath, stDatas)
	go db_analyze.AnalyzePostOperationData(user, password, ip, dbName, stDatas)
	// 保存文件
	db_analyze.SaveData(stDatas)
}
