package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"net/http"
)

// DeviceUnbind 解绑设备接口
// 只需要验证 token，不需要 TOTP 验证
func DeviceUnbind(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "未登录")
		return
	}

	// 查找并删除绑定记录（用户唯一，直接按用户名删除）
	result := database.DB.Where("user_name = ?", userInfo.UserName).
		Delete(&system.UserDevice{})
	if result.Error != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "解绑失败")
		return
	}

	if result.RowsAffected == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "未找到绑定记录")
		return
	}

	middleware.Logger.Infof("[设备解绑] 用户 %s 解绑设备成功", userInfo.UserName)

	middleware.ResponseSuccess(w, "解绑成功", nil)
}
