package export

import (
	"cmdb/models/sql"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// SqlExportGenerateDownload 生成临时下载链接
func SqlExportGenerateDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "仅支持POST请求")
		return
	}

	// 验证JWT Token（通过中间件自动验证）
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 解析请求体
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求体失败")
		return
	}
	defer r.Body.Close()

	if req.ID == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "导出ID不能为空")
		return
	}

	// 查询导出记录
	var exportInfo sql.SQLExport
	if err := database.DB.Where("id = ?", req.ID).First(&exportInfo).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "导出记录不存在")
		return
	}

	// 权限检查：只有提交人、审批人、审核人、执行人可以生成下载链接
	if !canDownload(userInfo.UserID, &exportInfo) {
		middleware.ResponseError(w, http.StatusForbidden, "无权生成此导出的下载链接")
		return
	}

	// 检查导出状态（StatusCompleted = 4 表示执行成功）
	if exportInfo.Status != sql.StatusCompleted {
		middleware.ResponseError(w, http.StatusBadRequest, "导出未完成，无法生成下载链接")
		return
	}

	// 检查文件路径
	if exportInfo.FilePath == "" {
		middleware.ResponseError(w, http.StatusNotFound, "文件不存在")
		return
	}

	// 检查文件是否存在于磁盘
	if _, err := os.Stat(exportInfo.FilePath); os.IsNotExist(err) {
		middleware.ResponseError(w, http.StatusNotFound, "文件已被删除")
		return
	}

	// 生成随机下载码
	downloadCode := uuid.New().String()

	// 存储到Redis（只存储ID）
	redisKey := fmt.Sprintf("%s%s", ExportDownloadCodeKeyPrefix, downloadCode)
	expireTime := 600 * time.Second // 600秒过期（10分钟）

	middleware.Logger.Infof("准备存储下载码到Redis: key=%s, ID=%d", redisKey, req.ID)

	if err := database.RDB.Set(r.Context(), redisKey, fmt.Sprintf("%d", req.ID), expireTime).Err(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "生成下载链接失败")
		middleware.Logger.Errorf("存储下载码到Redis失败: %v", err)
		return
	}

	middleware.Logger.Infof("成功存储下载码到Redis: key=%s", redisKey)

	// 从数据库获取平台地址
	var setting system.Setting
	downloadURL := "/api/sql/export/download?code=" + downloadCode
	if err := database.DB.First(&setting).Error; err == nil && setting.SystemUrl != "" {
		downloadURL = setting.SystemUrl + "/api/sql/export/download?code=" + downloadCode
	}

	expireAt := time.Now().Add(expireTime)

	middleware.ResponseSuccess(w, "生成下载链接成功", map[string]interface{}{
		"downloadUrl":  downloadURL,
		"downloadCode": downloadCode,
		"expireIn":     60, // 60秒（1分钟）
		"expireAt":     expireAt.Format("2006-01-02 15:04:05"),
	})

	middleware.Logger.Infof("生成临时下载链接: ID=%d, code=%s, user=%s", req.ID, downloadCode, userInfo.UserName)
}

// GenerateDownloadLink 生成下载链接（内部函数，供自动生成使用）
// 返回: downloadURL, downloadCode, expireAt, error
func GenerateDownloadLink(exportID int64) (string, string, time.Time, error) {
	// 查询导出记录
	var exportInfo sql.SQLExport
	if err := database.DB.Where("id = ?", exportID).First(&exportInfo).Error; err != nil {
		return "", "", time.Time{}, fmt.Errorf("导出记录不存在: %v", err)
	}

	// 检查导出状态
	if exportInfo.Status != sql.StatusCompleted {
		return "", "", time.Time{}, fmt.Errorf("导出未完成，无法生成下载链接")
	}

	// 检查文件路径
	if exportInfo.FilePath == "" {
		return "", "", time.Time{}, fmt.Errorf("文件不存在")
	}

	// 检查文件是否存在于磁盘
	if _, err := os.Stat(exportInfo.FilePath); os.IsNotExist(err) {
		return "", "", time.Time{}, fmt.Errorf("文件已被删除")
	}

	// 生成随机下载码
	downloadCode := uuid.New().String()

	// 存储到Redis（只存储ID）
	redisKey := fmt.Sprintf("%s%s", ExportDownloadCodeKeyPrefix, downloadCode)
	expireTime := 600 * time.Second // 600秒过期（10分钟）

	middleware.Logger.Infof("自动生成下载码并存储到Redis: key=%s, ID=%d", redisKey, exportID)

	ctx := context.Background()
	if err := database.RDB.Set(ctx, redisKey, fmt.Sprintf("%d", exportID), expireTime).Err(); err != nil {
		middleware.Logger.Errorf("存储下载码到Redis失败: %v", err)
		return "", "", time.Time{}, fmt.Errorf("生成下载链接失败: %v", err)
	}

	middleware.Logger.Infof("成功存储下载码到Redis: key=%s", redisKey)

	// 从数据库获取平台地址
	var setting system.Setting
	downloadURL := "/api/sql/export/download?code=" + downloadCode
	if err := database.DB.First(&setting).Error; err == nil && setting.SystemUrl != "" {
		downloadURL = setting.SystemUrl + "/api/sql/export/download?code=" + downloadCode
	}

	expireAt := time.Now().Add(expireTime)

	middleware.Logger.Infof("自动生成临时下载链接成功: ID=%d, code=%s, expireAt=%s",
		exportID, downloadCode, expireAt.Format("2006-01-02 15:04:05"))

	return downloadURL, downloadCode, expireAt, nil
}
