package export

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// Redis键前缀
const ExportDownloadCodeKeyPrefix = "cmdb:export:download:"

// SqlExportDownload 下载导出文件（支持两种方式：1.直接下载 2.通过下载码下载）
func SqlExportDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 从URL参数获取下载码（临时链接方式）
	downloadCode := r.URL.Query().Get("code")
	if downloadCode != "" {
		// 方式2：通过下载码下载（走白名单，无需JWT）
		downloadByCode(w, r, downloadCode)
		return
	}

	// 方式1：直接下载（需要JWT鉴权）
	downloadDirectly(w, r)
}

// downloadDirectly 直接下载（需要JWT鉴权）
func downloadDirectly(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "无法获取用户信息")
		return
	}

	// 获取流程ID
	exportIDStr := r.URL.Query().Get("id")
	if exportIDStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少流程ID参数")
		return
	}

	var exportID int64
	if _, err := parseExportID(exportIDStr, &exportID); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "流程ID格式错误")
		return
	}

	// 查询导出记录
	var exportInfo sql.SQLExport
	if err := database.DB.Where("id = ?", exportID).First(&exportInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.ResponseError(w, http.StatusNotFound, "流程不存在")
		} else {
			middleware.ResponseError(w, http.StatusInternalServerError, "查询流程失败")
		}
		return
	}

	// 权限检查：只有提交人、审批人、审核人、执行人可以下载
	if !canDownload(userInfo.UserID, &exportInfo) {
		middleware.ResponseError(w, http.StatusForbidden, "无权下载此文件")
		return
	}

	// 检查文件是否存在
	if exportInfo.FilePath == "" {
		middleware.ResponseError(w, http.StatusNotFound, "文件不存在")
		return
	}

	// 检查下载次数
	if exportInfo.DownloadLimit > 0 && exportInfo.DownloadCount >= exportInfo.DownloadLimit {
		middleware.ResponseError(w, http.StatusForbidden, "已达到下载次数限制")
		return
	}

	// 检查文件是否存在于磁盘
	if _, err := os.Stat(exportInfo.FilePath); os.IsNotExist(err) {
		middleware.ResponseError(w, http.StatusNotFound, "文件已被删除")
		return
	}

	// 更新下载次数
	database.DB.Model(&sql.SQLExport{}).Where("id = ?", exportID).
		Update("download_count", gorm.Expr("download_count + 1"))

	// 检查文件是否存在于磁盘
	fileInfo, err := os.Stat(exportInfo.FilePath)
	if os.IsNotExist(err) {
		middleware.ResponseError(w, http.StatusNotFound, "文件已被删除")
		return
	}

	// 打开文件
	file, err := os.Open(exportInfo.FilePath)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "打开文件失败")
		return
	}
	defer file.Close()

	// 设置响应头
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(exportInfo.FileName)))
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// 使用流式传输，避免大文件占用内存
	http.ServeContent(w, r, exportInfo.FileName, fileInfo.ModTime(), file)

	middleware.Logger.Infof("文件下载成功，流程ID: %d, 用户: %s, 文件大小: %d", exportID, userInfo.UserName, fileInfo.Size())
}

// downloadByCode 通过下载码下载（无需JWT）
func downloadByCode(w http.ResponseWriter, r *http.Request, downloadCode string) {
	middleware.Logger.Infof("开始下载导出文件: code=%s", downloadCode)

	// 从Redis获取exportID
	redisKey := fmt.Sprintf("%s%s", ExportDownloadCodeKeyPrefix, downloadCode)
	middleware.Logger.Infof("查询Redis: key=%s", redisKey)

	exportIDStr, err := database.RDB.Get(r.Context(), redisKey).Result()
	if err != nil {
		middleware.Logger.Errorf("从Redis获取下载码失败: key=%s, error=%v", redisKey, err)
		middleware.ResponseError(w, http.StatusUnauthorized, "下载链接已过期或无效")
		return
	}

	// 解析exportID
	var exportID int64
	if _, err := fmt.Sscanf(exportIDStr, "%d", &exportID); err != nil {
		middleware.Logger.Errorf("解析exportID失败: %s, error=%v", exportIDStr, err)
		middleware.ResponseError(w, http.StatusInternalServerError, "解析下载链接失败")
		return
	}

	middleware.Logger.Infof("从Redis获取到exportID: %d", exportID)

	// 获取导出记录
	var exportInfo sql.SQLExport
	if err := database.DB.Where("id = ?", exportID).First(&exportInfo).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "导出记录不存在")
		return
	}

	// 检查导出状态（StatusCompleted = 4 表示执行成功）
	if exportInfo.Status != sql.StatusCompleted {
		middleware.ResponseError(w, http.StatusBadRequest, "导出未完成，无法下载")
		return
	}

	// 检查文件路径
	if exportInfo.FilePath == "" {
		middleware.ResponseError(w, http.StatusNotFound, "文件不存在")
		return
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(exportInfo.FilePath)
	if os.IsNotExist(err) {
		middleware.ResponseError(w, http.StatusNotFound, "文件已被删除")
		return
	}

	// 打开文件
	file, err := os.Open(exportInfo.FilePath)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "打开文件失败")
		return
	}
	defer file.Close()

	// 设置响应头
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(exportInfo.FileName)))
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// 使用流式传输，避免大文件占用内存
	http.ServeContent(w, r, exportInfo.FileName, fileInfo.ModTime(), file)

	middleware.Logger.Infof("下载导出文件: exportID=%d, file=%s, size=%d, code=%s", exportID, exportInfo.FileName, fileInfo.Size(), downloadCode)
}

// canDownload 检查用户是否有权下载
func canDownload(userID int64, export *sql.SQLExport) bool {
	return userID == export.SubmitterID ||
		userID == export.ApplyID ||
		userID == export.ReviewerID ||
		userID == export.ExecutorID
}

// parseExportID 解析导出ID
func parseExportID(s string, id *int64) (bool, error) {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, errors.New("invalid id")
		}
		n = n*10 + int64(c-'0')
	}
	*id = n
	return true, nil
}
