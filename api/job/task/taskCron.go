package task

import (
	"cmdb/pkg/middleware"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"net/http"
	"time"
)

// CronRequest cron表达式验证请求
type CronRequest struct {
	Cron string `json:"cron"` // cron表达式
}

// CronResponse cron表达式验证响应
type CronResponse struct {
	IsValid bool     `json:"is_valid"` // 是否有效
	Message string   `json:"message"`  // 错误信息
	NextRun []string `json:"next_run"` // 未来5次执行时间
}

// JobCron 验证cron表达式
func JobCron(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 解析请求数据
	var req CronRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 参数验证
	if req.Cron == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少cron表达式")
		return
	}

	// 验证cron表达式
	response := CronResponse{
		IsValid: true,
		NextRun: make([]string, 0, 5),
	}

	// 使用cron库解析表达式
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(req.Cron)
	if err != nil {
		response.IsValid = false
		response.Message = "无效的cron表达式: " + err.Error()
		middleware.ResponseSuccess(w, "验证完成", response)
		return
	}

	// 计算未来5次执行时间
	now := time.Now()
	next := now
	for i := 0; i < 5; i++ {
		next = schedule.Next(next)
		response.NextRun = append(response.NextRun, next.Format("2006-01-02 15:04:05"))
	}

	middleware.ResponseSuccess(w, "验证完成", response)
}
