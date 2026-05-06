package databi

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"encoding/json"
	"net/http"
)

// DataBIRefresh 触发异步刷新任务
// POST /api/sql/databi/refresh?project=xxx
func DataBIRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 获取项目参数
	project := r.URL.Query().Get("project")
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	// 验证项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !isSuperAdmin && !utils.Contains(projects, project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	ctx := context.Background()

	// 尝试获取刷新锁，防止并发刷新
	acquired, err := TryAcquireRefreshLock(ctx, project)
	if err != nil {
		middleware.Logger.Errorf("获取刷新锁失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "获取刷新锁失败")
		return
	}

	if !acquired {
		// 已有刷新任务在进行中
		middleware.ResponseSuccess(w, "刷新任务已在进行中", map[string]interface{}{
			"project": project,
			"status":  "refreshing",
		})
		return
	}

	// 启动后台异步刷新任务
	go executeRefreshTask(project)

	// 立即返回
	middleware.ResponseSuccess(w, "刷新任务已启动", map[string]interface{}{
		"project": project,
		"status":  "started",
	})
}

// executeRefreshTask 执行刷新任务（后台运行）
func executeRefreshTask(project string) {
	ctx := context.Background()

	// 确保任务结束时释放锁
	defer func() {
		if err := ReleaseRefreshLock(ctx, project); err != nil {
			middleware.Logger.Errorf("[DataBI刷新] 释放锁失败 [项目:%s] - %v", project, err)
		}
	}()

	middleware.Logger.Infof("[DataBI刷新] 开始刷新任务 [项目:%s]", project)

	// 步骤1：设置状态为 refreshing（刷新插件缓存）
	middleware.Logger.Infof("[DataBI刷新] 步骤1：设置状态为 refreshing [项目:%s]", project)
	if err := SaveRefreshStatus(ctx, project, RefreshStatus{
		Status:   "refreshing",
		Message:  "正在刷新插件缓存...",
		Progress: 20,
	}); err != nil {
		middleware.Logger.Errorf("[DataBI刷新] 保存状态失败: %v", err)
	} else {
		middleware.Logger.Infof("[DataBI刷新] 状态已保存到Redis [项目:%s, 状态:refreshing]", project)
	}

	// 调用插件 loadTable 接口，触发刷新插件内存缓存（等待完成，但不关心返回结果）
	middleware.Logger.Infof("[DataBI刷新] 步骤2：调用loadTable [项目:%s]", project)
	pluginPath := "/proxy/databi-plugs/spark/common/loadTable"
	middleware.CallAgent(project, pluginPath) // 等待调用完成
	middleware.Logger.Infof("[DataBI刷新] loadTable调用完成 [项目:%s]", project)

	// 步骤2：设置状态为 loading（加载表列表）
	middleware.Logger.Infof("[DataBI刷新] 步骤3：设置状态为 loading [项目:%s]", project)
	if err := SaveRefreshStatus(ctx, project, RefreshStatus{
		Status:   "loading",
		Message:  "正在加载表列表...",
		Progress: 60,
	}); err != nil {
		middleware.Logger.Errorf("[DataBI刷新] 保存状态失败: %v", err)
	} else {
		middleware.Logger.Infof("[DataBI刷新] 状态已保存到Redis [项目:%s, 状态:loading]", project)
	}

	// 调用插件 tables 接口，获取最新表列表
	middleware.Logger.Infof("[DataBI刷新] 步骤4：调用tables [项目:%s]", project)
	tablesPath := "/proxy/databi-plugs/spark/common/tables"
	tablesData, err := middleware.CallAgent(project, tablesPath)

	// 解析 tables 响应（尽力解析，即使失败也继续）
	var tablesResp struct {
		Code int                 `json:"code"`
		Msg  string              `json:"msg"`
		Data map[string][]string `json:"data"`
	}

	// 如果调用失败或解析失败，使用空数据
	if err != nil {
		middleware.Logger.Warnf("[DataBI刷新] 调用tables失败 [项目:%s] - %v，使用空数据", project, err)
		tablesResp.Data = make(map[string][]string)
	} else if err := json.Unmarshal(tablesData, &tablesResp); err != nil {
		middleware.Logger.Warnf("[DataBI刷新] 解析tables响应失败 [项目:%s] - %v，使用空数据", project, err)
		tablesResp.Data = make(map[string][]string)
	} else if tablesResp.Code != 200 {
		middleware.Logger.Warnf("[DataBI刷新] tables返回错误 [项目:%s] - %s，使用空数据", project, tablesResp.Msg)
		tablesResp.Data = make(map[string][]string)
	}

	middleware.Logger.Infof("[DataBI刷新] tables调用完成 [项目:%s, 数据库数量:%d]", project, len(tablesResp.Data))

	// 步骤3：保存表列表到 Redis 缓存
	middleware.Logger.Infof("[DataBI刷新] 步骤5：保存表列表到Redis [项目:%s, 数据库数量:%d]", project, len(tablesResp.Data))
	if err := SaveTablesToCache(ctx, project, tablesResp.Data); err != nil {
		middleware.Logger.Errorf("[DataBI刷新] 保存表列表到缓存失败 [项目:%s] - %v", project, err)
		// 不中断流程，继续设置成功状态
	} else {
		middleware.Logger.Infof("[DataBI刷新] 表列表已保存到Redis [项目:%s]", project)
	}

	// 步骤4：设置状态为 success，并包含表数据
	middleware.Logger.Infof("[DataBI刷新] 步骤6：设置状态为 success [项目:%s]", project)
	if err := SaveRefreshStatus(ctx, project, RefreshStatus{
		Status:   "success",
		Message:  "刷新完成",
		Progress: 100,
		Tables:   tablesResp.Data,
	}); err != nil {
		middleware.Logger.Errorf("[DataBI刷新] 保存成功状态失败: %v", err)
	} else {
		middleware.Logger.Infof("[DataBI刷新] 成功状态已保存到Redis [项目:%s]", project)
	}

	middleware.Logger.Infof("[DataBI刷新] 刷新任务完成 [项目:%s]", project)
}
