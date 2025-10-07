package internal

import "time"

const dateTimeLayout = "2006_01_02_15_04_05"

var beijingTZ = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("Beijing", 8*60*60)
	}
	return loc
}()

func timeToBeijingTimeStr(t time.Time) string {
	return t.In(beijingTZ).Format(dateTimeLayout)
}

func beijingTimeStrToTime(s string) time.Time {
	t, _ := time.ParseInLocation(dateTimeLayout, s, beijingTZ)
	return t
}
