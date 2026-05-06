package view

import (
	"cmdb/models/elfk"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"net/http"
	"time"
)

// CreateView 创建视图
func CreateView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var view elfk.View
	if err := json.NewDecoder(r.Body).Decode(&view); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	// 验证必填字段
	if view.Name == "" || view.Project == "" || view.IndexPattern == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "名称、项目、索引模式和时间字段不能为空")
		return
	}

	// 设置创建和更新时间
	now := time.Now()
	view.CreateTime = now
	view.UpdateTime = now

	// 保存到数据库
	if err := database.DB.Create(&view).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建视图失败")
		return
	}
	err := UpdateViewRedisCache(view.ID)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "视图缓存写入redis失败")
		return
	}
	middleware.ResponseSuccess(w, "创建视图成功", view)
}
