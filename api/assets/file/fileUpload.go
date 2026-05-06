package file

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// FileUpload 上传文件到指定目录
// @Summary 上传文件
// @Description 上传文件到指定项目的目录，支持批量上传和目录结构保留
// @Tags 资产管理
// @Accept multipart/form-data
// @Produce json
// @Param project query string true "项目标识"
// @Param key formData string true "目录标识(假根)"
// @Param path formData string false "相对路径(已废弃，请使用filePath)"
// @Param file formData file true "上传的文件(可多个)"
// @Param filePath formData string false "文件的相对路径(可多个，与file一一对应，如: folder/subfolder/)"
// @Success 200 {object} map[string]interface{}
// @Router /api/assets/file/upload [post]
func FileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 解析multipart表单，限制最大内存为32MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析表单失败: "+err.Error())
		return
	}

	// 获取参数
	project := r.FormValue("project")
	if project == "" {
		project = r.URL.Query().Get("project")
	}
	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}

	key := r.FormValue("key")
	path := r.FormValue("path")
	if key == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "key参数不能为空")
		return
	}

	// 获取Agent地址
	agentURL, err := system.GetProjectAgentURL(project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("获取项目Agent地址失败: %v", err))
		return
	}

	// 获取文件和路径
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "未找到上传文件")
		return
	}

	filePaths := r.MultipartForm.Value["filePath"]

	middleware.Logger.Infof("接收到文件上传请求 - 项目: %s, key: %s, 文件数量: %d, 路径数量: %d",
		project, key, len(files), len(filePaths))

	// 构建文件名到路径的映射（通过分析前端请求顺序）
	// 前端发送顺序: file1, filePath1, file2, filePath2, ...
	// Go的MultipartForm会分别收集，但保持各自的顺序
	// 所以 files[i] 对应 filePaths[i]

	// 构建转发请求
	var bodyBuffer bytes.Buffer
	multipartWriter := multipart.NewWriter(&bodyBuffer)

	var totalSize int64
	for i, fileHeader := range files {
		filename := fileHeader.Filename

		// 获取对应的filePath（如果有）
		var filePath string
		if i < len(filePaths) {
			filePath = filePaths[i]
		}

		if filePath != "" {
			middleware.Logger.Infof("[文件 %d] 文件名: %s, 指定路径: %s, 大小: %d",
				i+1, filename, filePath, fileHeader.Size)
		} else {
			middleware.Logger.Infof("[文件 %d] 文件名: %s, 路径: (根目录), 大小: %d",
				i+1, filename, fileHeader.Size)
		}

		// 打开文件
		file, err := fileHeader.Open()
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError,
				fmt.Sprintf("打开文件 %s 失败: %v", filename, err))
			return
		}

		// 写入filePath字段（即使为空也要写，保持顺序对应）
		if err := multipartWriter.WriteField("filePath", filePath); err != nil {
			file.Close()
			middleware.ResponseError(w, http.StatusInternalServerError,
				fmt.Sprintf("写入filePath字段失败: %v", err))
			return
		}

		// 写入file字段
		part, err := multipartWriter.CreateFormFile("file", filename)
		if err != nil {
			file.Close()
			middleware.ResponseError(w, http.StatusInternalServerError,
				fmt.Sprintf("创建文件字段 %s 失败: %v", filename, err))
			return
		}

		written, err := io.Copy(part, file)
		file.Close()
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError,
				fmt.Sprintf("复制文件 %s 内容失败: %v", filename, err))
			return
		}
		totalSize += written
	}

	if err := multipartWriter.Close(); err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError,
			"关闭multipart writer失败: "+err.Error())
		return
	}

	// 发送请求到Agent
	reqURL := fmt.Sprintf("%s/proxy/file-plugs/api/upload?key=%s&path=%s", agentURL, key, path)
	contentType := multipartWriter.FormDataContentType()

	middleware.Logger.Infof("===== 文件上传请求组装信息 =====")
	middleware.Logger.Infof("目标URL: %s", reqURL)
	middleware.Logger.Infof("Content-Type: %s", contentType)
	middleware.Logger.Infof("文件数量: %d", len(files))
	middleware.Logger.Infof("文件总大小: %d bytes, 请求体大小: %d bytes", totalSize, bodyBuffer.Len())
	middleware.Logger.Infof("================================")

	req, err := http.NewRequest("POST", reqURL, &bodyBuffer)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建请求失败: "+err.Error())
		return
	}
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError,
			fmt.Sprintf("请求Agent失败: %v", err))
		return
	}
	defer resp.Body.Close()

	// 处理响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取Agent响应失败")
		return
	}

	cfg := config.GetConfig()
	taskSalt := cfg.Agent.Secret
	middleware.Logger.Infof("使用taskSalt解密代理响应，taskSalt长度: %d", len(taskSalt))
	middleware.Logger.Infof("代理返回原始响应，状态码: %d, 响应体: %s", resp.StatusCode, string(respBody))

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		middleware.Logger.Errorf("解析响应失败: %v, 原始响应: %s", err, string(respBody))
		middleware.ResponseError(w, http.StatusInternalServerError, "解析响应失败")
		return
	}

	middleware.Logger.Infof("解析后的响应结构: %+v", result)

	if encryptedRespData, ok := result["data"].(string); ok && encryptedRespData != "" {
		middleware.Logger.Infof("开始解密代理响应data字段，加密数据长度: %d", len(encryptedRespData))

		decompressedData, err := middleware.DecryptAndDecompress(encryptedRespData, taskSalt)
		if err != nil {
			middleware.Logger.Errorf("解密解压失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "解密解压失败: "+err.Error())
			return
		}

		middleware.Logger.Infof("解密解压成功，解密后数据: %s", string(decompressedData))

		var uploadResult interface{}
		if err := json.Unmarshal(decompressedData, &uploadResult); err != nil {
			middleware.Logger.Errorf("解析上传结果数据失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "解析上传结果数据失败")
			return
		}

		middleware.Logger.Infof("文件上传成功: %+v", uploadResult)
		middleware.ResponseSuccess(w, "文件上传成功", uploadResult)
	} else {
		middleware.Logger.Errorf("代理返回错误响应，状态码: %d, 无有效data字段", resp.StatusCode)

		errorMsg := ""
		if msg, ok := result["msg"].(string); ok {
			errorMsg = msg
		} else if errStr, ok := result["error"].(string); ok {
			errorMsg = errStr
		} else if message, ok := result["message"].(string); ok {
			errorMsg = message
		} else {
			errorMsg = fmt.Sprintf("未知错误，状态码: %d, 完整响应: %+v", resp.StatusCode, result)
		}

		middleware.ResponseError(w, resp.StatusCode, errorMsg)
	}
}
