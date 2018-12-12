package db_analyze

import (
	"fmt"
	"time"
)

const (
	TempTableName       = "raiing_tcms_temp_data"
	UserTableName       = "raiing_tcms_user"
	EventTableName      = "raiing_tcms_event_data"
	TemperatureInterval = 4 // 温度间隔
	STDataTempFileName  = "temp_file.csv"
)

func GetTimeNow() string {
	timeNow := time.Now()
	year := timeNow.Year()
	month := timeNow.Month()
	day := timeNow.Day()
	hour := timeNow.Hour()
	minute := timeNow.Minute()
	//second := timeNow.Second()
	timeStr := fmt.Sprintf("%d%02d%02d_%02d%02d", year, month, day, hour, minute)
	return timeStr
}
