package time

import (
	"strconv"
	"time"
)

// 日期差计算
// 年月日计算
func SubDate(date1, date2 time.Time) string {
	var y, m, d int
	y = date1.Year() - date2.Year()
	if date1.Month() < date2.Month() {
		y--
		m = 12 - int(date2.Month()) + int(date1.Month())
	} else {
		m = int(date1.Month()) - int(date2.Month())
	}
	// 天数模糊计算
	if date1.Day() < date2.Day() {
		m--
		//闰年,29天
		day := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

		if date2.Year()%4 == 0 && date2.Year()%100 != 0 || date2.Year()%400 == 0 {
			d = day[date2.Month()-1] + 1 - date2.Day() + date1.Day()
		} else {
			d = day[date2.Month()-1] - date2.Day() + date1.Day()
		}
	} else {
		d = date1.Day() - date2.Day()
	}
	return strconv.Itoa(y) + "年" + strconv.Itoa(m) + "月" + strconv.Itoa(d) + "日"
}