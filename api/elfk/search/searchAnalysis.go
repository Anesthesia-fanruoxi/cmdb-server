package search

import (
	"cmdb/api/task"
	"cmdb/models/elfk"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnalysisRequest 分析请求结构体
type AnalysisRequest struct {
	Qid            string `json:"qid"`            // 查询ID
	Field          string `json:"field"`          // 需要分析的字段名
	StartDelimiter string `json:"startDelimiter"` // 开始分隔符(当直接指定字段无法获取到值时使用)
	EndDelimiter   string `json:"endDelimiter"`   // 结束分隔符(当直接指定字段无法获取到值时使用)
	SortBy         string `json:"sortBy"`         // 排序方式：count-按计数排序，value-按值排序
	LogType        string `json:"log_type"`       // 日志类型：elfk 或 sls
	Count          int    `json:"count"`          // 需要分析的数据量

	// 原始查询参数
	Project      string `json:"project"`       // 项目标识
	ViewName     string `json:"view_name"`     // 视图名称
	IndexPattern string `json:"index_pattern"` // 索引模式
	TimeField    string `json:"time_field"`    // 时间字段
	StartTime    string `json:"start_time"`    // 查询开始时间
	EndTime      string `json:"end_time"`      // 查询结束时间
	Keyword      string `json:"keyword"`       // 关键字
}

// AnalysisResult 分析结果结构体
type AnalysisResult struct {
	Values map[string]int `json:"values"` // 值和计数的映射
}

// GetQueryInfoByID 根据查询ID获取查询信息
func GetQueryInfoByID(ctx context.Context, queryID string) (project, viewName, indexPattern, timeField, timeFormat, startTime, endTime, keyword string, err error) {
	// 优先从数据库审计日志获取查询信息（避免Redis超时问题）
	var searchAudit elfk.SysEsLog
	if err := database.DB.Where("query_id = ?", queryID).First(&searchAudit).Error; err == nil {
		// 从审计记录获取基本信息
		project = searchAudit.Project
		viewName = searchAudit.ViewName
		indexPattern = searchAudit.QIndexPattern
		timeField = searchAudit.QTimeField
		startTime = searchAudit.QStartTime
		endTime = searchAudit.QEndTime
		keyword = searchAudit.QKeyword

		// 从视图表中获取时间格式
		if viewName != "" && project != "" {
			var view elfk.View
			if err := database.DB.Where("name = ? AND project = ?", viewName, project).First(&view).Error; err == nil {
				timeFormat = view.TimeFormat
			}
		}

		middleware.Logger.Infof("从数据库获取查询信息成功: project=%s, indexPattern=%s", project, indexPattern)
		return project, viewName, indexPattern, timeField, timeFormat, startTime, endTime, keyword, nil
	}

	// 如果数据库中没有找到，尝试从Redis获取（备选方案）
	if database.RDB != nil {
		// 设置较短的超时时间避免长时间等待
		timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		cachedResult, found := GetESResultFromCacheByID(timeoutCtx, database.RDB, queryID)
		if found {
			project = cachedResult.Project
			indexPattern = cachedResult.IndexName
			middleware.Logger.Infof("从Redis缓存获取查询信息成功: project=%s, indexPattern=%s", project, indexPattern)
			return project, viewName, indexPattern, timeField, timeFormat, startTime, endTime, keyword, nil
		}
	}

	return "", "", "", "", "", "", "", "", fmt.Errorf("未找到查询信息，queryID: %s", queryID)
}

// RegisterAnalysisExecutor 注册数据分析执行器
func RegisterAnalysisExecutor() {
	if task.GlobalTaskManager == nil {
		middleware.Logger.Error("TaskManager未初始化，无法注册数据分析执行器")
		return
	}

	task.GlobalTaskManager.RegisterExecutor("analysis", &DataAnalysisExecutor{})
	middleware.Logger.Info("数据分析执行器注册成功，类型: analysis")
}

// DataAnalysisExecutor 数据分析任务执行器
type DataAnalysisExecutor struct{}

// Execute 实现 TaskExecutor 接口
func (e *DataAnalysisExecutor) Execute(ctx context.Context, t *task.Task) error {
	middleware.Logger.Infof("开始执行数据分析任务: %s", t.ID)

	// 步骤1开始（初始化）
	now := time.Now()
	t.Setup.Setup1.Status = "running"
	t.Setup.Setup1.Description = "初始化分析任务"
	t.Setup.Setup1.StartTime = now
	t.Status = task.TaskStatusRunning
	t.Progress = 0

	// 1. 验证任务参数
	if t.Params == nil {
		t.Setup.Setup1.Status = "failed"
		t.Setup.Setup1.EndTime = time.Now()
		t.Setup.Setup1.Duration = t.Setup.Setup1.EndTime.Sub(t.Setup.Setup1.StartTime).Milliseconds()
		t.Setup.Setup1.Description = "任务参数为空(nil)"
		t.Status = task.TaskStatusFailed
		t.ErrorMessage = "任务参数为空(nil)"
		return fmt.Errorf("任务参数为空(nil)")
	}

	paramBytes := []byte(t.Params)
	if len(paramBytes) == 0 {
		t.Setup.Setup1.Status = "failed"
		t.Setup.Setup1.EndTime = time.Now()
		t.Setup.Setup1.Duration = t.Setup.Setup1.EndTime.Sub(t.Setup.Setup1.StartTime).Milliseconds()
		t.Setup.Setup1.Description = "任务参数为空字符串"
		t.Status = task.TaskStatusFailed
		t.ErrorMessage = "任务参数为空字符串"
		return fmt.Errorf("任务参数为空字符串")
	}

	// 步骤1完成
	t.Setup.Setup1.Status = "success"
	t.Setup.Setup1.EndTime = time.Now()
	t.Setup.Setup1.Duration = t.Setup.Setup1.EndTime.Sub(t.Setup.Setup1.StartTime).Milliseconds()
	t.Setup.Setup1.Description = "初始化分析任务完成"
	t.Progress = 25

	// 步骤2开始（准备数据）
	t.Setup.Setup2.Status = "running"
	t.Setup.Setup2.Description = "准备分析数据"
	t.Setup.Setup2.StartTime = time.Now()

	// 2. 解析任务参数
	var req AnalysisRequest
	if err := json.Unmarshal(paramBytes, &req); err != nil {
		t.Setup.Setup2.Status = "failed"
		t.Setup.Setup2.EndTime = time.Now()
		t.Setup.Setup2.Duration = t.Setup.Setup2.EndTime.Sub(t.Setup.Setup2.StartTime).Milliseconds()
		t.Setup.Setup2.Description = fmt.Sprintf("解析任务参数失败: %v", err)
		t.Status = task.TaskStatusFailed
		t.ErrorMessage = fmt.Sprintf("解析任务参数失败: %v", err)
		return fmt.Errorf("解析任务参数失败: %v, 原始参数: %s", err, string(paramBytes))
	}

	// 3. 验证必要字段
	if req.Qid == "" {
		t.Setup.Setup2.Status = "failed"
		t.Setup.Setup2.EndTime = time.Now()
		t.Setup.Setup2.Duration = t.Setup.Setup2.EndTime.Sub(t.Setup.Setup2.StartTime).Milliseconds()
		t.Setup.Setup2.Description = "查询ID不能为空"
		t.Status = task.TaskStatusFailed
		t.ErrorMessage = "查询ID不能为空"
		return fmt.Errorf("查询ID不能为空")
	}

	if req.Field == "" {
		t.Setup.Setup2.Status = "failed"
		t.Setup.Setup2.EndTime = time.Now()
		t.Setup.Setup2.Duration = t.Setup.Setup2.EndTime.Sub(t.Setup.Setup2.StartTime).Milliseconds()
		t.Setup.Setup2.Description = "分析字段不能为空"
		t.Status = task.TaskStatusFailed
		t.ErrorMessage = "分析字段不能为空"
		return fmt.Errorf("分析字段不能为空")
	}

	// 设置默认值
	if req.Count <= 0 {
		req.Count = 1000 // 默认分析1000条记录
	}

	if req.LogType == "" {
		req.LogType = "elfk" // 默认使用ELFK
	}

	// 步骤2完成
	t.Setup.Setup2.Status = "success"
	t.Setup.Setup2.EndTime = time.Now()
	t.Setup.Setup2.Duration = t.Setup.Setup2.EndTime.Sub(t.Setup.Setup2.StartTime).Milliseconds()
	t.Setup.Setup2.Description = "准备分析数据完成"
	t.Progress = 50

	// 步骤3开始（执行分析）
	t.Setup.Setup3.Status = "running"
	t.Setup.Setup3.Description = "执行数据分析"
	t.Setup.Setup3.StartTime = time.Now()

	// 4. 根据日志类型执行不同的分析逻辑
	var result *AnalysisResult
	var err error

	switch req.LogType {
	case "elfk":
		result, err = analyzeElfkDataAsync(ctx, req, t)
	case "sls":
		result, err = analyzeSlsDataAsync(ctx, req, t)
	default:
		t.Setup.Setup3.Status = "failed"
		t.Setup.Setup3.EndTime = time.Now()
		t.Setup.Setup3.Duration = t.Setup.Setup3.EndTime.Sub(t.Setup.Setup3.StartTime).Milliseconds()
		t.Setup.Setup3.Description = fmt.Sprintf("不支持的日志类型: %s", req.LogType)
		t.Status = task.TaskStatusFailed
		t.ErrorMessage = fmt.Sprintf("不支持的日志类型: %s", req.LogType)
		return fmt.Errorf("不支持的日志类型: %s", req.LogType)
	}

	if err != nil {
		t.Setup.Setup3.Status = "failed"
		t.Setup.Setup3.EndTime = time.Now()
		t.Setup.Setup3.Duration = t.Setup.Setup3.EndTime.Sub(t.Setup.Setup3.StartTime).Milliseconds()
		t.Setup.Setup3.Description = fmt.Sprintf("执行分析失败: %v", err)
		t.Status = task.TaskStatusFailed
		t.ErrorMessage = fmt.Sprintf("执行分析失败: %v", err)
		return fmt.Errorf("执行分析失败: %v", err)
	}

	// 步骤3完成
	t.Setup.Setup3.Status = "success"
	t.Setup.Setup3.EndTime = time.Now()
	t.Setup.Setup3.Duration = t.Setup.Setup3.EndTime.Sub(t.Setup.Setup3.StartTime).Milliseconds()
	t.Setup.Setup3.Description = "执行数据分析完成"
	t.Progress = 75

	// 步骤4开始（保存结果）
	t.Setup.Setup4.Status = "running"
	t.Setup.Setup4.Description = "保存分析结果"
	t.Setup.Setup4.StartTime = time.Now()

	// 5. 生成Excel文件保存完整结果
	filePath, fileName, fileSize, err := saveAnalysisToExcel(t.UserID, t.ID, result)
	if err != nil {
		t.Setup.Setup4.Status = "failed"
		t.Setup.Setup4.EndTime = time.Now()
		t.Setup.Setup4.Duration = t.Setup.Setup4.EndTime.Sub(t.Setup.Setup4.StartTime).Milliseconds()
		t.Setup.Setup4.Description = fmt.Sprintf("生成Excel文件失败: %v", err)
		t.Status = task.TaskStatusFailed
		t.ErrorMessage = fmt.Sprintf("生成Excel文件失败: %v", err)
		return fmt.Errorf("生成Excel文件失败: %v", err)
	}

	// 6. Redis只存前100条预览数据
	previewResult := getPreviewResult(result, 100)
	previewJSON, err := json.Marshal(previewResult)
	if err != nil {
		middleware.Logger.Warnf("序列化预览数据失败: %v", err)
	} else {
		if err := task.GlobalTaskManager.UpdateTaskResult(t.ID, string(previewJSON)); err != nil {
			middleware.Logger.Warnf("保存预览数据到Redis失败: %v", err)
		}
	}

	// 7. 更新MySQL任务记录
	totalCount := len(result.Values)
	if err := task.UpdateTaskSuccess(t.ID, filePath, fileName, fileSize, totalCount); err != nil {
		middleware.Logger.Warnf("更新MySQL任务记录失败: %v", err)
	}

	// 步骤4完成
	t.Setup.Setup4.Status = "success"
	t.Setup.Setup4.EndTime = time.Now()
	t.Setup.Setup4.Duration = t.Setup.Setup4.EndTime.Sub(t.Setup.Setup4.StartTime).Milliseconds()
	t.Setup.Setup4.Description = "保存分析结果完成"
	t.Status = task.TaskStatusSuccess
	t.Progress = 100

	middleware.Logger.Infof("数据分析任务执行完成: %s, 文件: %s, 总数: %d", t.ID, filePath, totalCount)
	return nil
}

// ESSearchAnalysis HTTP处理函数 - 数据分析
func ESSearchAnalysis(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 记录开始时间，用于计算执行时间
	startTime := time.Now()

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	// 记录原始请求体
	middleware.Logger.Infof("收到分析请求，原始数据: %s", string(body))

	// 解析JSON
	var req AnalysisRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("解析JSON失败: %v", err))
		return
	}

	// 记录解析后的请求对象
	middleware.Logger.Infof("解析后的请求对象: %+v", req)

	// 验证必填字段
	if req.Qid == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "查询ID不能为空")
		return
	}

	if req.Field == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "字段名不能为空")
		return
	}

	if req.LogType == "" {
		req.LogType = "elfk" // 默认使用ELFK
	}

	// 获取关联查询的完整信息
	project, viewName, indexPattern, timeField, _, queryStartTime, queryEndTime, keyword, err := GetQueryInfoByID(r.Context(), req.Qid)
	if err != nil {
		middleware.Logger.Warnf("获取查询信息失败: %v", err)
		// 尝试从Redis缓存中获取部分信息
		if cachedResult, found := GetESResultFromCacheByID(r.Context(), database.RDB, req.Qid); found {
			project = cachedResult.Project
			indexPattern = cachedResult.IndexName
		} else {
			middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取查询信息失败: %v", err))
			return
		}
	}

	// 添加原始查询参数
	req.Project = project
	req.ViewName = viewName
	req.IndexPattern = indexPattern
	req.TimeField = timeField
	req.StartTime = queryStartTime
	req.EndTime = queryEndTime
	req.Keyword = keyword

	// 将解析后的请求对象重新序列化为JSON
	taskParams, err := json.Marshal(req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("序列化任务参数失败: %v", err))
		return
	}

	// 记录任务参数
	middleware.Logger.Infof("创建任务的参数: %s", string(taskParams))

	// 创建异步分析任务
	analysisTask, err := task.GlobalTaskManager.CreateTask(
		"analysis",
		json.RawMessage(taskParams), // 确保参数类型为 json.RawMessage
		userInfo.UserName,           // 使用实际用户名，而不是"system"
		userInfo.UserID,             // 使用实际用户ID，而不是0
	)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("创建分析任务失败: %v", err))
		return
	}

	// 计算查询耗时
	duration := time.Since(startTime).Milliseconds()

	// 记录审计日志
	go RecordESSearchAudit(
		r,                 // 请求对象
		req.Qid,           // 查询ID
		project,           // 项目标识
		viewName,          // 视图名称
		indexPattern,      // 索引模式
		timeField,         // 时间字段
		queryStartTime,    // 开始时间
		queryEndTime,      // 结束时间
		keyword,           // 搜索关键词
		duration,          // 查询耗时
		0,                 // 文档数量（分析任务暂不统计）
		OperationAnalysis, // 操作类型为分析
		0,                 // 页码（分析操作不涉及分页）
		http.StatusOK,     // 响应状态码
		nil,               // 错误信息
	)

	// 返回任务ID
	middleware.ResponseSuccess(w, "分析任务已创建", map[string]interface{}{
		"task_id": analysisTask.ID,
		"message": "分析任务已创建，请使用任务ID查询结果",
	})
}

// analyzeElfkDataAsync 异步分析ELFK数据
func analyzeElfkDataAsync(_ context.Context, req AnalysisRequest, t *task.Task) (*AnalysisResult, error) {
	// 直接调用已有的分析函数
	return analyzeElfkData(req, t)
}

// analyzeSlsDataAsync 异步分析SLS数据
func analyzeSlsDataAsync(_ context.Context, req AnalysisRequest, t *task.Task) (*AnalysisResult, error) {
	// 直接调用已有的分析函数
	return analyzeSlsData(req, t)
}
