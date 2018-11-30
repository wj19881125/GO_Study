package main

import "db_analyze"

func main() {
	// 传入参数
	// 1. 手术室所有压缩包解压以后的路径
	rawDataPath := "X:/GO/raw_data"
	// 2. TCMS服务器IP,数据库用户名和密码
	user := "root"
	password := "123456"
	ip := "127.0.0.1:3306"
	dbName := "raiing_tcms_v6_temp"
	// 3. 保存文件名
	// 4. 保存文件路径

	stDatas := make(chan db_analyze.StData, 20)

	go db_analyze.AnalyzeInOperationData(rawDataPath, stDatas)
	go db_analyze.AnalyzePostOperationData(user, password, ip, dbName, stDatas)
	// 保存文件
	db_analyze.SaveData(stDatas)
}
