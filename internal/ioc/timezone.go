package ioc

import "time"

func TimezoneInit() {
	// 设置时区为北京时间
	time.LoadLocation("Asia/Shanghai")
}
