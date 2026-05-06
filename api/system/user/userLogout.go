package user

import (
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
	"strings"
)

// UserLogout 用户登出
func UserLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "未找到用户信息")
		return
	}

	// 获取token
	token := r.Header.Get("Authorization")
	if token == "" {
		middleware.ResponseError(w, http.StatusUnauthorized, "未提供token")
		return
	}

	// 去掉Bearer前缀
	token = strings.TrimPrefix(token, "Bearer ")

	// 使token失效
	if err := utils.InvalidateToken(userInfo.UserID); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "登出失败")
		return
	}

	middleware.ResponseSuccess(w, "登出成功", nil)
}
