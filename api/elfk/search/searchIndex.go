package search

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// EsInfo ES信息查询接口
// 传 project 和 index：获取指定索引的字段结构
func EsInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取参数
	project := r.URL.Query().Get("project")
	index := r.URL.Query().Get("index")
	logType := r.URL.Query().Get("log_type") // 获取log_type参数

	// 检查项目参数
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "项目标识不能为空")
		return
	}

	// 检查索引参数
	if index == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "索引参数不能为空")
		return
	}

	// 获取项目权限
	projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 如果不是超级管理员且没有权限访问该项目，返回无权限错误
	if !isSuperAdmin && !utils.Contains(projects, project) {
		middleware.ResponseError(w, http.StatusForbidden, "无权访问该项目")
		return
	}

	// 根据log_type参数选择不同的处理方式
	if logType == "sls" {
		// 使用SLS API查询索引信息
		GetSLSIndexInfo(w, project, index)
	} else {
		// 默认使用ELFK查询
		GetELFKIndexInfo(w, r, project, index)
	}
}
