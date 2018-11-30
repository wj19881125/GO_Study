package db_analyze

import (
	"fmt"
	"time"
)

const (
	TempTableName       = "raiing_tcms_temp_data"
	B2W_TABLE_NAME      = "raiing_tcms_b2w"
	UserTableName       = "raiing_tcms_user"
	EventTableName      = "raiing_tcms_event_data"
	TemperatureInterval = 4 // 温度间隔
)

func GetTimeNow() string {
	timeNow := time.Now()
	year := timeNow.Year()
	month := timeNow.Month()
	day := timeNow.Day()
	hour := timeNow.Hour()
	minute := timeNow.Minute()
	second := timeNow.Second()
	timeStr := fmt.Sprintf("%d%d%d_%d%d_%d", year, month, day, hour, minute, second)
	return timeStr
}
