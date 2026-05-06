package appupdate

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"net/http"
)

const versionKey = middleware.KeyPathAppVersion

// VersionInfo 版本信息
type VersionInfo struct {
	Version     string `json:"version"`     // 版本号
	Description string `json:"description"` // 版本说明
}

// GetVersion 获取版本信息
// GET /api/app/version/list
func GetVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持GET请求")
		return
	}

	ctx := context.Background()
	data, err := database.RDB.Get(ctx, versionKey).Result()
	if err != nil {
		// 没有数据返回空
		middleware.ResponseSuccess(w, "获取成功", VersionInfo{})
		return
	}

	var info VersionInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解析版本信息失败")
		return
	}

	middleware.ResponseSuccess(w, "获取成功", info)
}

// UpdateVersion 更新版本信息
// POST /api/system/appupdate/version
func UpdateVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	var req VersionInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数格式错误")
		return
	}

	if req.Version == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "版本号不能为空")
		return
	}

	data, err := json.Marshal(req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "序列化失败")
		return
	}

	ctx := context.Background()
	if err := database.RDB.Set(ctx, versionKey, data, 0).Err(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "保存失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", req)
}
