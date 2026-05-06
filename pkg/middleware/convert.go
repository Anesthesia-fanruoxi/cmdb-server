package middleware

import (
	"time"
)

// ToInt64 将任意数值类型转换为int64
func ToInt64(v interface{}) int64 {
	switch v := v.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

// FromInt64 将int64转换为指定类型
func FromInt64[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](v int64) T {
	return T(v)
}

// FormatTimeToChina 将时间转换为中国格式 (yyyy-MM-dd HH:mm:ss)
func FormatTimeToChina(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t = t.In(loc)
	return t.Format("2006-01-02 15:04:05")
}

// TimeToMilliseconds 将时间转换为毫秒时间戳
func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// GetChinaTime 获取中国时区的当前时间
func GetChinaTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc)
}
