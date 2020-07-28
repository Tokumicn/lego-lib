package utime

import "time"

// Millisec 返回当前unix时间戳 ms
func Millisec() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Duration 数值转换为time.Duration
func Duration(millisec int32) time.Duration {
	return time.Duration(millisec) * time.Millisecond
}

// UtcZero 当天utc零点(北京8点)的时间戳
func UtcZero() (int64, error) {
	timess := time.Now().Format("2006-01-02")
	t, err := time.Parse("2006-01-02", timess)
	return t.UnixNano() / 1e6, err
}
