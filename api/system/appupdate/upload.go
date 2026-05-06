package appupdate

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cmdb/pkg/middleware"
)

const uploadDir = "/mnt/client"

// Upload 上传安装包（支持 msi/dmg/exe）
func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 限制上传大小 500MB
	r.Body = http.MaxBytesReader(w, r.Body, 500<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请选择文件")
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			middleware.Logger.Errorf("关闭上传文件失败: %v", closeErr)
		}
	}()

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".msi" && ext != ".dmg" {
		middleware.ResponseError(w, http.StatusBadRequest, "仅支持 msi、dmg 文件")
		return
	}

	// 确保目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建目录失败")
		return
	}

	// 保存文件（自动覆盖）
	destPath := filepath.Join(uploadDir, header.Filename)
	out, err := os.Create(destPath)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建文件失败")
		return
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			middleware.Logger.Errorf("关闭目标文件失败: %v", closeErr)
		}
	}()

	if _, err := io.Copy(out, file); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "保存文件失败")
		return
	}

	middleware.ResponseSuccess(w, "上传成功", map[string]string{
		"filename": header.Filename,
		"path":     destPath,
	})
}
