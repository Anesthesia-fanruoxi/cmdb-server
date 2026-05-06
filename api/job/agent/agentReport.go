package agent

import (
	"cmdb/config"
	"cmdb/models/job"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AgentReportRequest Agent上报执行结果请求
type AgentReportRequest struct {
	JobID      int64     `json:"job_id"`      // 任务ID
	Project    string    `json:"project"`     // 项目标识
	TaskName   string    `json:"task_name"`   // 任务名称
	TaskKey    string    `json:"task_key"`    // 任务标识符
	ExecLog    string    `json:"exec_log"`    // 执行日志(脚本输出)
	ExecStatus int       `json:"exec_status"` // 执行状态 0:执行中 1:成功 2:失败
	Result     string    `json:"result"`      // 执行结果
	ErrorMsg   string    `json:"error_msg"`   // 错误信息
	StartTime  time.Time `json:"start_time"`  // 开始时间
	EndTime    time.Time `json:"end_time"`    // 结束时间
	Duration   int       `json:"duration"`    // 执行时长(秒)
}

// AgentReportResult Agent上报任务执行结果
func AgentReportResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 读取原始请求数据
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求数据失败")
		return
	}
	defer r.Body.Close()

	middleware.Logger.Infof("接收到上报请求，数据长度: %d", len(body))
	middleware.Logger.Infof("原始请求体: %s", string(body))

	// 从请求中获取data字段
	var requestData struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(body, &requestData); err != nil {
		middleware.Logger.Errorf("JSON解析失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据格式错误")
		return
	}

	// 参数验证
	if requestData.Data == "" {
		middleware.Logger.Error("data字段为空")
		middleware.ResponseError(w, http.StatusBadRequest, "缺少必要参数")
		return
	}

	middleware.Logger.Infof("获取到加密数据，长度: %d", len(requestData.Data))

	// 获取任务加密盐
	cfg := config.GetConfig()
	taskSalt := cfg.Agent.Secret
	if taskSalt == "" {
		middleware.ResponseError(w, http.StatusInternalServerError, "Agent加密密钥未配置")
		return
	}

	// 解密并解压数据
	decompressedData, err := middleware.DecryptAndDecompress(requestData.Data, taskSalt)
	if err != nil {
		middleware.Logger.Errorf("解密解压失败: %v", err)
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("解密解压失败: %v", err))
		return
	}

	middleware.Logger.Infof("解密成功，解密后数据长度: %d", len(decompressedData))
	middleware.Logger.Infof("解密后数据内容: %s", string(decompressedData))

	// 解析解密后的请求数据
	var req AgentReportRequest
	if err := json.Unmarshal(decompressedData, &req); err != nil {
		middleware.Logger.Errorf("请求数据解析失败: %v, 数据内容: %s", err, string(decompressedData))
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	middleware.Logger.Infof("解密后的上报数据: JobID=%d, Status=%d, Duration=%d", req.JobID, req.ExecStatus, req.Duration)

	// 参数验证
	if req.JobID == 0 {
		middleware.Logger.Errorf("上报失败: job_id为空")
		middleware.ResponseError(w, http.StatusBadRequest, "job_id不能为空")
		return
	}

	middleware.Logger.Infof("开始处理上报 [JobID: %d, Status: %d, Duration: %d]", req.JobID, req.ExecStatus, req.Duration)

	// 查询执行记录
	var execution job.JobExecution
	if err := database.DB.First(&execution, req.JobID).Error; err != nil {
		middleware.Logger.Errorf("查询执行记录失败 [JobID: %d]: %v", req.JobID, err)
		middleware.ResponseError(w, http.StatusNotFound, "执行记录不存在")
		return
	}

	middleware.Logger.Infof("找到执行记录 [JobID: %d, 当前状态: %d]", req.JobID, execution.ExecStatus)

	// 更新执行记录
	updates := map[string]interface{}{
		"exec_log":    req.ExecLog,
		"exec_status": req.ExecStatus,
		"result":      req.Result,
		"error_msg":   req.ErrorMsg,
		"end_time":    &req.EndTime,
		"duration":    req.Duration,
	}

	if err := database.DB.Model(&execution).Updates(updates).Error; err != nil {
		middleware.Logger.Errorf("更新执行记录失败 [JobID: %d]: %v", req.JobID, err)
		middleware.ResponseError(w, http.StatusInternalServerError, "更新执行记录失败")
		return
	}

	middleware.Logger.Infof("执行记录更新成功 [JobID: %d, 新状态: %d]", req.JobID, req.ExecStatus)
	middleware.ResponseSuccess(w, "上报成功", nil)
}
