package file

import (
	"cmdb/config"
	"cmdb/models/system"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// FileList 获取目录文件列表
// @Summary 获取目录文件列表
// @Description 根据project和key获取指定目录下的文件列表，支持分页和搜索
// @Tags 资产管理
// @Accept json
// @Produce json
// @Param project query string true "项目标识"
// @Param key query string true "目录标识(假根)"
// @Param path query string false "相对路径"
// @Param page query int false "页码(默认1)"
// @Param search query string false "搜索关键词"
// @Param sort query string false "排序方式: name(默认)、time(按时间倒序)、size(按大小倒序)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/assets/file/list [get]
func FileList(w http.ResponseWriter, r *http.Request) {
	// 1. 获取参数
	project := r.URL.Query().Get("project")
	key := r.URL.Query().Get("key")
	path := r.URL.Query().Get("path")
	page := r.URL.Query().Get("page")
	search := r.URL.Query().Get("search")
	sort := r.URL.Query().Get("sort")

	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}
	if key == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "key参数不能为空")
		return
	}

	// 2. 获取项目信息（包含Agent地址和三方域名）
	projectInfo, err := system.GetProjectInfo(project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("获取项目信息失败: %v", err))
		return
	}

	agentURL := projectInfo.AgentURL
	if agentURL == "" {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("项目 %s 未配置agent地址", project))
		return
	}

	// 3. 调用Agent接口
	// Agent接口：/api/proxy/file-plugs/api/list?key=xxx&path=xxx&page=xxx&search=xxx
	// key用于选择假根，path用于访问子目录，page用于分页，search用于搜索
	reqURL := fmt.Sprintf("%s/proxy/file-plugs/api/list?key=%s&path=%s&page=%s&search=%s&sort=%s", agentURL, key, path, page, search, sort)

	// 创建请求
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建请求失败")
		return
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("请求Agent失败: %v", err))
		return
	}
	defer resp.Body.Close()

	// 4. 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "读取Agent响应失败")
		return
	}

	// 获取任务加密盐
	cfg := config.GetConfig()
	taskSalt := cfg.Agent.Secret
	middleware.Logger.Infof("使用taskSalt解密代理响应，taskSalt长度: %d", len(taskSalt))

	// 记录原始响应
	middleware.Logger.Infof("代理返回原始响应，状态码: %d, 响应体: %s", resp.StatusCode, string(respBody))

	// 解析响应体
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		middleware.Logger.Errorf("解析响应失败: %v, 原始响应: %s", err, string(respBody))
		middleware.ResponseError(w, http.StatusInternalServerError, "解析响应失败")
		return
	}

	middleware.Logger.Infof("解析后的响应结构: %+v", result)

	// 检查是否包含data字段（无论状态码如何）
	if encryptedRespData, ok := result["data"].(string); ok && encryptedRespData != "" {
		middleware.Logger.Infof("开始解密代理响应data字段，加密数据长度: %d", len(encryptedRespData))

		// 解密并解压数据 - 这里解密的是代理的响应体
		decompressedData, err := middleware.DecryptAndDecompress(encryptedRespData, taskSalt)
		if err != nil {
			middleware.Logger.Errorf("解密解压失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "解密解压失败: "+err.Error())
			return
		}

		middleware.Logger.Infof("解密解压成功，解密后数据: %s", string(decompressedData))

		// 解密后得到的是文件列表数据
		var fileListData map[string]interface{}
		if err := json.Unmarshal(decompressedData, &fileListData); err != nil {
			middleware.Logger.Errorf("解析文件列表数据失败: %v", err)
			middleware.ResponseError(w, http.StatusInternalServerError, "解析文件列表数据失败")
			return
		}

		// 处理数据重组，添加访问地址
		// 只有key为html(协议)或verify(校验文件)时需要添加url
		if data, ok := fileListData["data"].(map[string]interface{}); ok {
			// 获取文件列表
			if files, ok := data["files"].([]interface{}); ok {
				newFiles := make([]map[string]interface{}, 0, len(files))
				apiDomain := strings.TrimSuffix(projectInfo.ApiDomain, "/")

				for _, f := range files {
					if fileMap, ok := f.(map[string]interface{}); ok {
						// 获取文件名
						name, _ := fileMap["name"].(string)

						// 根据key类型生成访问地址
						var accessURL string
						if key == "verify" {
							// 校验文件: api_domain/filename
							if apiDomain != "" {
								accessURL = fmt.Sprintf("%s/%s", apiDomain, name)
							}
						} else if key == "html" {
							// 协议文件: api_domain/html/[path/]filename
							if apiDomain != "" {
								urlPath := "html"
								if path != "" {
									urlPath = fmt.Sprintf("html/%s", strings.Trim(path, "/"))
								}
								accessURL = fmt.Sprintf("%s/%s/%s", apiDomain, urlPath, name)
							}
						}

						// 只有当生成了URL时才添加字段
						if accessURL != "" {
							fileMap["url"] = accessURL
						}

						newFiles = append(newFiles, fileMap)
					}
				}
				// 更新文件列表
				data["files"] = newFiles
			}
		}

		middleware.Logger.Infof("文件列表解析成功: %+v", fileListData)

		// 返回解密后的文件列表
		middleware.ResponseSuccess(w, "获取文件列表成功", fileListData)
	} else {
		// 如果没有data字段或data为空，说明是真正的错误响应
		middleware.Logger.Errorf("代理返回错误响应，状态码: %d, 无有效data字段", resp.StatusCode)

		// 提取错误信息
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
