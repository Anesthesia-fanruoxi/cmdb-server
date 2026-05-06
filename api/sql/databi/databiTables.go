package databi

import (
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DataBITables 通过 SSE 实时推送 Redis 中的刷新状态和表数据
// GET /api/sql/databi/tables?project=xxx&token=xxx
func DataBITables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	// 获取 URL 参数
	query := r.URL.Query()

	// 从 URL 获取 token
	token := query.Get("token")
	if token == "" {
		middleware.ResponseError(w, http.StatusUnauthorized, "未提供token")
		return
	}

	// 验证 token
	jwtConfig := config.GetConfig().JWT
	userInfo, err := utils.ParseToken(token, jwtConfig.Secret)
	if err != nil {
		middleware.ResponseError(w, http.StatusUnauthorized, "token无效: "+err.Error())
		return
	}

	// 获取项目参数
	project := query.Get("project")
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	// 获取用户角色级别
	var role system.Role
	if err := database.DB.First(&role, userInfo.RoleID).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取角色信息失败")
		return
	}

	// 创建权限检查实例
	auth := utils.ProjectAuth(
		uint(userInfo.UserID),
		uint(userInfo.DeptID),
		uint(userInfo.RoleID),
		role.Level,
		userInfo.OnlineAssets,
		userInfo.TestAssets,
		r.URL.Path,
	)

	// 获取可访问项目列表
	projects, err := auth.GetAccessibleProjects("prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "查询可访问项目失败")
		return
	}

	// 判断是否为超级管理员
	isSuperAdmin := (role.Level == 0)

	if !isSuperAdmin && !utils.Contains(projects, project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := r.Context()
	flusher, ok := w.(http.Flusher)
	if !ok {
		middleware.Logger.Error("不支持流式响应")
		middleware.ResponseError(w, http.StatusInternalServerError, "不支持流式响应")
		return
	}

	middleware.Logger.Infof("[DataBI SSE] 客户端连接 [项目:%s]", project)

	// 发送初始状态（连接建立时发送一次）
	sendInitialStatus(w, flusher, project)

	// 轮询 Redis 检测状态变化（事件驱动）
	ticker := time.NewTicker(500 * time.Millisecond) // 每500ms检查一次
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute) // 最长连接5分钟

	var lastStatusJSON string // 记录上次状态的 JSON，用于检测变化

	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			middleware.Logger.Infof("[DataBI SSE] 客户端断开连接 [项目:%s]", project)
			return

		case <-timeout:
			// 超时，关闭连接
			middleware.Logger.Infof("[DataBI SSE] 连接超时 [项目:%s]", project)
			return

		case <-ticker.C:
			// 从 Redis 获取刷新状态
			status, err := GetRefreshStatus(context.Background(), project)

			var currentStatusJSON string

			if err != nil {
				if err == database.RedisNil {
					// 没有刷新任务，检查是否有缓存
					cache, cacheErr := GetTablesFromCache(context.Background(), project)
					if cacheErr == nil {
						// 有缓存数据
						cacheData := map[string]interface{}{
							"status":      "cached",
							"message":     "使用缓存数据",
							"tables":      cache.Tables,
							"update_time": cache.UpdateTime,
						}
						cacheJSON, _ := json.Marshal(cacheData)
						currentStatusJSON = string(cacheJSON)
					} else {
						// 无缓存，空闲状态
						idleData := map[string]interface{}{
							"status":  "idle",
							"message": "请点击刷新按钮",
						}
						idleJSON, _ := json.Marshal(idleData)
						currentStatusJSON = string(idleJSON)
					}
				} else {
					middleware.Logger.Errorf("[DataBI SSE] 获取状态失败: %v", err)
					continue
				}
			} else {
				// 有刷新状态
				statusJSON, _ := json.Marshal(status)
				currentStatusJSON = string(statusJSON)
			}

			// 检测状态是否发生变化
			if currentStatusJSON != lastStatusJSON {
				middleware.Logger.Infof("[DataBI SSE] 检测到状态变化 [项目:%s]", project)

				// 解析并发送变化后的状态
				var statusData map[string]interface{}
				json.Unmarshal([]byte(currentStatusJSON), &statusData)

				eventType := statusData["status"].(string)
				middleware.Logger.Infof("[DataBI SSE] 发送状态更新 [项目:%s, 状态:%s]", project, eventType)
				sendSSEEvent(w, flusher, eventType, statusData)

				// 更新上次状态
				lastStatusJSON = currentStatusJSON
			}
		}
	}
}

// sendInitialStatus 发送初始状态（连接建立时）
func sendInitialStatus(w http.ResponseWriter, flusher http.Flusher, project string) {
	// 检查是否有刷新任务
	status, err := GetRefreshStatus(context.Background(), project)
	if err == nil {
		// 有刷新任务，发送当前状态
		middleware.Logger.Infof("[DataBI SSE] 发送初始状态 [项目:%s, 状态:%s]", project, status.Status)
		sendSSEEvent(w, flusher, status.Status, status)
		return
	}

	// 没有刷新任务，检查缓存
	cache, cacheErr := GetTablesFromCache(context.Background(), project)
	if cacheErr == nil {
		// 有缓存，发送缓存数据
		middleware.Logger.Infof("[DataBI SSE] 发送初始缓存数据 [项目:%s]", project)
		sendSSEEvent(w, flusher, "cached", map[string]interface{}{
			"status":      "cached",
			"message":     "使用缓存数据",
			"tables":      cache.Tables,
			"update_time": cache.UpdateTime,
		})
		return
	}

	// 无缓存，发送空闲状态
	middleware.Logger.Infof("[DataBI SSE] 发送初始空闲状态 [项目:%s]", project)
	sendSSEEvent(w, flusher, "idle", map[string]interface{}{
		"status":  "idle",
		"message": "请点击刷新按钮",
	})
}

// sendSSEEvent 发送 SSE 事件
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data interface{}) {
	dataJSON, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", string(dataJSON))
	flusher.Flush()
	middleware.Logger.Infof("[DataBI SSE] 已发送事件 [类型:%s]", eventType)
}
