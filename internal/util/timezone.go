package util

import (
	"os"
	"time"
)

// 全局北京时区变量
var BeijingLocation *time.Location

func init() {
	var err error
	BeijingLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载时区失败，使用固定的UTC+8偏移量
		BeijingLocation = time.FixedZone("CST", 8*60*60)
	}
}

// TimezoneInit 初始化时区设置
func TimezoneInit() {
	// 设置时区环境变量为北京时间
	os.Setenv("TZ", "Asia/Shanghai")
	// 重新加载时区
	var err error
	BeijingLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		BeijingLocation = time.FixedZone("CST", 8*60*60)
	}
}

// GetBeijingTime 获取当前北京时间
func GetBeijingTime() time.Time {
	return time.Now().In(BeijingLocation)
}

// ToBeijingTime 将时间转换为北京时间
func ToBeijingTime(t time.Time) time.Time {
	return t.In(BeijingLocation)
}

// FormatBeijingTime 将时间格式化为北京时间字符串
func FormatBeijingTime(t time.Time, layout string) string {
	return t.In(BeijingLocation).Format(layout)
}
