package utils

import (
	"fmt"
	"strings"
)

// TranslateCron 将cron表达式翻译为人类可读的中文描述
func TranslateCron(cronExpr string) string {
	// 处理空表达式
	if cronExpr == "" {
		return "无调度计划"
	}

	// 分割cron表达式
	parts := strings.Fields(cronExpr)
	if len(parts) < 5 {
		return "无效的cron表达式"
	}

	// 提取各部分
	var second, minute, hour, day, month, week string
	if len(parts) == 6 {
		// 包含秒的cron表达式: 秒 分 时 日 月 周
		second, minute, hour, day, month, week = parts[0], parts[1], parts[2], parts[3], parts[4], parts[5]
	} else {
		// 不包含秒的cron表达式: 分 时 日 月 周
		second = "0"
		minute, hour, day, month, week = parts[0], parts[1], parts[2], parts[3], parts[4]
	}

	// 处理特殊情况 - 精确匹配常见模式

	// 每秒执行
	if second == "*" && minute == "*" && hour == "*" && day == "*" && month == "*" && (week == "*" || week == "?") {
		return "每秒执行"
	}

	// 每分钟执行
	if (second == "0" || second == "00") && minute == "*" && hour == "*" && day == "*" && month == "*" && (week == "*" || week == "?") {
		return "每分钟执行"
	}

	// 每小时执行
	if (second == "0" || second == "00") && (minute == "0" || minute == "00") && hour == "*" && day == "*" && month == "*" && (week == "*" || week == "?") {
		return "每小时执行"
	}

	// 每天执行
	if (second == "0" || second == "00") && (minute == "0" || minute == "00") && (hour == "0" || hour == "00") && day == "*" && month == "*" && (week == "*" || week == "?") {
		return "每天零点执行"
	}

	// 处理步长表达式

	// 秒级别的步长
	if strings.HasPrefix(second, "*/") && minute == "*" && hour == "*" && day == "*" && month == "*" && (week == "*" || week == "?") {
		stepStr := strings.TrimPrefix(second, "*/")
		return fmt.Sprintf("每隔%s秒执行", stepStr)
	}

	// 分钟级别的步长
	if (second == "0" || second == "00") && strings.HasPrefix(minute, "*/") && hour == "*" && day == "*" && month == "*" && (week == "*" || week == "?") {
		stepStr := strings.TrimPrefix(minute, "*/")
		return fmt.Sprintf("每隔%s分钟执行", stepStr)
	}

	// 小时级别的步长
	if (second == "0" || second == "00") && (minute == "0" || minute == "00") && strings.HasPrefix(hour, "*/") && day == "*" && month == "*" && (week == "*" || week == "?") {
		stepStr := strings.TrimPrefix(hour, "*/")
		return fmt.Sprintf("每隔%s小时执行", stepStr)
	}

	// 处理常见的时间点模式

	// 每天特定时间
	if (second == "0" || second == "00") && minute != "*" && hour != "*" && day == "*" && month == "*" && (week == "*" || week == "?") {
		return fmt.Sprintf("每天%s点%s分执行", hour, minute)
	}

	// 每小时特定分钟
	if (second == "0" || second == "00") && minute != "*" && hour == "*" && day == "*" && month == "*" && (week == "*" || week == "?") {
		return fmt.Sprintf("每小时第%s分执行", minute)
	}

	// 每分钟特定秒数
	if second != "*" && minute == "*" && hour == "*" && day == "*" && month == "*" && (week == "*" || week == "?") {
		return fmt.Sprintf("每分钟第%s秒执行", second)
	}

	// 每周特定时间
	if (second == "0" || second == "00") && minute != "*" && hour != "*" && day == "?" && month == "*" && week != "*" && week != "?" {
		weekDesc := translateWeek(week)
		return fmt.Sprintf("%s的%s点%s分执行", weekDesc, hour, minute)
	}

	// 每月特定日期和时间
	if (second == "0" || second == "00") && minute != "*" && hour != "*" && day != "*" && day != "?" && month == "*" && (week == "*" || week == "?") {
		return fmt.Sprintf("每月%s日%s点%s分执行", day, hour, minute)
	}

	// 构建通用描述
	var desc strings.Builder

	// 处理月份
	if month != "*" {
		desc.WriteString(translateMonth(month))
	} else {
		desc.WriteString("每月")
	}

	// 处理日期和星期
	if day != "*" && day != "?" {
		desc.WriteString(translateDay(day))
	} else if week != "*" && week != "?" {
		desc.WriteString(translateWeek(week))
	} else {
		desc.WriteString("每天")
	}

	// 处理小时
	if hour != "*" {
		desc.WriteString(fmt.Sprintf("%s点", translateHour(hour)))
	} else {
		desc.WriteString("每小时")
	}

	// 处理分钟
	if minute != "*" {
		desc.WriteString(fmt.Sprintf("%s分", translateMinute(minute)))
	} else if hour != "*" {
		desc.WriteString("每分钟")
	}

	// 处理秒
	if second != "*" && second != "0" && second != "00" {
		desc.WriteString(fmt.Sprintf("%s秒", translateSecond(second)))
	}

	desc.WriteString("执行")

	return desc.String()
}

// translateSecond 翻译秒表达式
func translateSecond(expr string) string {
	if expr == "0" || expr == "00" {
		return "0"
	}
	return translateTimeUnit(expr)
}

// translateMinute 翻译分钟表达式
func translateMinute(expr string) string {
	return translateTimeUnit(expr)
}

// translateHour 翻译小时表达式
func translateHour(expr string) string {
	return translateTimeUnit(expr)
}

// translateDay 翻译日期表达式
func translateDay(expr string) string {
	return translateTimeUnit(expr) + "日"
}

// translateMonth 翻译月份表达式
func translateMonth(expr string) string {
	// 替换月份名称
	expr = strings.ReplaceAll(expr, "JAN", "1")
	expr = strings.ReplaceAll(expr, "FEB", "2")
	expr = strings.ReplaceAll(expr, "MAR", "3")
	expr = strings.ReplaceAll(expr, "APR", "4")
	expr = strings.ReplaceAll(expr, "MAY", "5")
	expr = strings.ReplaceAll(expr, "JUN", "6")
	expr = strings.ReplaceAll(expr, "JUL", "7")
	expr = strings.ReplaceAll(expr, "AUG", "8")
	expr = strings.ReplaceAll(expr, "SEP", "9")
	expr = strings.ReplaceAll(expr, "OCT", "10")
	expr = strings.ReplaceAll(expr, "NOV", "11")
	expr = strings.ReplaceAll(expr, "DEC", "12")

	return translateTimeUnit(expr) + "月"
}

// translateWeek 翻译星期表达式
func translateWeek(expr string) string {
	// 替换星期名称
	expr = strings.ReplaceAll(expr, "SUN", "0")
	expr = strings.ReplaceAll(expr, "MON", "1")
	expr = strings.ReplaceAll(expr, "TUE", "2")
	expr = strings.ReplaceAll(expr, "WED", "3")
	expr = strings.ReplaceAll(expr, "THU", "4")
	expr = strings.ReplaceAll(expr, "FRI", "5")
	expr = strings.ReplaceAll(expr, "SAT", "6")

	// 将数字转换为中文星期
	weekMap := map[string]string{
		"0": "星期日",
		"7": "星期日", // 有些系统用7表示星期日
		"1": "星期一",
		"2": "星期二",
		"3": "星期三",
		"4": "星期四",
		"5": "星期五",
		"6": "星期六",
	}

	// 处理范围、列表和步长
	if strings.Contains(expr, "-") {
		// 范围: 1-5
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			start, ok1 := weekMap[parts[0]]
			end, ok2 := weekMap[parts[1]]
			if ok1 && ok2 {
				return fmt.Sprintf("从%s到%s", start, end)
			}
		}
	} else if strings.Contains(expr, ",") {
		// 列表: 1,3,5
		parts := strings.Split(expr, ",")
		var days []string
		for _, part := range parts {
			if day, ok := weekMap[part]; ok {
				days = append(days, day)
			}
		}
		return strings.Join(days, "、")
	} else if strings.Contains(expr, "/") {
		// 步长: */2
		parts := strings.Split(expr, "/")
		if len(parts) == 2 {
			if parts[0] == "*" {
				return fmt.Sprintf("每隔%s天", parts[1])
			}
		}
	} else {
		// 单个值
		if day, ok := weekMap[expr]; ok {
			return day
		}
	}

	return "每周"
}

// translateTimeUnit 通用时间单位翻译
func translateTimeUnit(expr string) string {
	// 处理通配符
	if expr == "*" {
		return "每"
	}

	// 处理范围
	if strings.Contains(expr, "-") {
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			return fmt.Sprintf("%s到%s", parts[0], parts[1])
		}
	}

	// 处理列表
	if strings.Contains(expr, ",") {
		return strings.ReplaceAll(expr, ",", "、")
	}

	// 处理步长
	if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		if len(parts) == 2 {
			if parts[0] == "*" {
				return fmt.Sprintf("每隔%s", parts[1])
			}
			return fmt.Sprintf("从%s开始每隔%s", parts[0], parts[1])
		}
	}

	// 单个值，直接返回
	return expr
}
