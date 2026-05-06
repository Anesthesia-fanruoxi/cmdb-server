package proUpdate

import (
	"cmdb/models/assets"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ProUpdateListDetail 获取某个项目的所有发版记录（SSE实时推送）
// GET /api/assets/proUpdate/listDetail
func ProUpdateListDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 获取项目参数
	project := r.URL.Query().Get("project")
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目参数不能为空")
		return
	}

	// 获取类型参数（web=前端，空=后端）
	requestType := r.URL.Query().Get("type")

	// 查询项目字典表，获取元数据
	var projectDict system.ProjectDict
	if err := database.DB.Where("project = ?", project).First(&projectDict).Error; err != nil {
		middleware.Logger.Errorf("查询项目字典失败: %v", err)
		middleware.ResponseError(w, http.StatusNotFound, "项目不存在")
		return
	}

	// 根据类型确定 Git 地址和工具
	var gitURL string
	var tool string
	if requestType == "web" {
		gitURL = projectDict.GitVue
		tool = projectDict.FrontendTool
	} else {
		gitURL = projectDict.GitBackend
		tool = projectDict.BackendTool
	}

	// 构建项目元数据（不变的信息）
	projectInfo := assets.ProjectInfo{
		Project:      project,
		ProjectName:  projectDict.ProjectName,
		GitURL:       gitURL,
		Tool:         tool,
		UpdateFeishu: projectDict.UpdateFeishu,
		NotifyFeishu: projectDict.NotifyFeishu,
	}

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no") // 禁用nginx缓冲

	// 获取Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		middleware.ResponseError(w, http.StatusInternalServerError, "服务器不支持SSE")
		return
	}

	middleware.Logger.Infof("SSE连接建立 - Project: %s, Type: %s", project, requestType)

	// 使用context监听客户端断开
	ctx := r.Context()

	// 每2秒查询一次数据库
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastHash string

	// 立即发送一次数据
	if data, hash, err := fetchProjectRecords(project, requestType, projectInfo); err == nil {
		sendSSEData(w, flusher, data)
		lastHash = hash
	}

	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			middleware.Logger.Infof("SSE连接断开 - Project: %s", project)
			return
		case <-ticker.C:
			// 查询数据库
			data, hash, err := fetchProjectRecords(project, requestType, projectInfo)
			if err != nil {
				middleware.Logger.Errorf("SSE获取数据失败: %v", err)
				sendSSEError(w, flusher, err.Error())
				continue
			}
			// 只有数据变化时才推送
			if hash != lastHash {
				sendSSEData(w, flusher, data)
				lastHash = hash
			}
		}
	}
}

// fetchProjectRecords 获取项目发版记录（近30天）
// project: 项目名称，requestType: 类型(web=前端，空=后端)
func fetchProjectRecords(project string, requestType string, projectInfo assets.ProjectInfo) (*assets.ProUpdateListDetailResponse, string, error) {
	// 计算30天前的时间
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// 查询项目发版记录（根据 project 和 type 筛选）
	var records []assets.ReleaseRecord
	query := database.DB.Where("project = ? AND started_at >= ?", project, thirtyDaysAgo)
	if requestType == "web" {
		// 前端项目
		query = query.Where("type = ?", "web")
	} else {
		// 后端项目（type 为空、double、single）
		query = query.Where("type != ? OR type = ''", "web")
	}
	if err := query.Order("started_at DESC").Find(&records).Error; err != nil {
		return nil, "", err
	}

	// 构建响应
	response := &assets.ProUpdateListDetailResponse{
		ProjectInfo:  projectInfo,
		TotalUpdates: len(records),
		LastUpdate:   "-",
		Records:      make([]assets.UpdateRecordInfo, 0, len(records)),
	}

	// 如果有记录，设置最后更新时间
	if len(records) > 0 && !records[0].StartedAt.IsZero() {
		response.LastUpdate = middleware.FormatTimeToChina(records[0].StartedAt)
	}

	// 添加所有记录
	for _, record := range records {
		startedAtStr := "-"
		completedAtStr := "-"

		if !record.StartedAt.IsZero() {
			startedAtStr = middleware.FormatTimeToChina(record.StartedAt)
		}

		if !record.CompletedAt.IsZero() {
			completedAtStr = middleware.FormatTimeToChina(record.CompletedAt)
		}

		response.Records = append(response.Records, assets.UpdateRecordInfo{
			ID:          record.ID,
			TaskID:      record.TaskID,
			Type:        record.Type,
			Status:      record.Status,
			Description: record.Description,
			StartedAt:   startedAtStr,
			CompletedAt: completedAtStr,
		})
	}

	// 生成hash用于变化检测
	hashData, _ := json.Marshal(response)
	hash := fmt.Sprintf("%x", len(hashData))

	return response, hash, nil
}

// sendSSEData 发送SSE数据
func sendSSEData(w http.ResponseWriter, flusher http.Flusher, data *assets.ProUpdateListDetailResponse) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: data\ndata: %s\n\n", jsonData)
	flusher.Flush()
}

// sendSSEError 发送SSE错误
func sendSSEError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", errMsg)
	flusher.Flush()
}
