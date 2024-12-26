package test

import (
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

// IterationRequest 迭代请求参数
type IterationRequest struct {
	GitURL      string `json:"git_url" binding:"required"`      // Git仓库地址
	GitBranch   string `json:"git_branch" binding:"required"`   // Git分支
	JavaVersion string `json:"java_version" binding:"required"` // Java版本
	Namespace   string `json:"namespace" binding:"required"`    // 命名空间
}

// StartIteration 开始迭代操作
func StartIteration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 解析请求参数
	var req IterationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "参数解析失败")
		return
	}

	// 参数验证
	if req.GitURL == "" || req.GitBranch == "" || req.JavaVersion == "" || req.Namespace == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "参数不能为空")
		return
	}

	// 获取当前用户名
	userName := r.Header.Get("X-User-Name")
	if userName == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "用户名不能为空")
		return
	}

}
