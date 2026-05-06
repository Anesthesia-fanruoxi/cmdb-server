package search

import (
	"cmdb/pkg/middleware"
	"fmt"
	"net/http"
	"strings"

	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

// GetSLSIndexInfo 使用SLS API获取索引信息
func GetSLSIndexInfo(w http.ResponseWriter, project, indexPattern string) {
	// 使用SLSManager获取对应项目的客户端
	client, err := GlobalSLSManager.GetClient(project)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取SLS客户端失败: %v", err))
		return
	}

	// 处理索引名称(通常格式为"project-app*")
	logstore := indexPattern
	if strings.HasSuffix(logstore, "*") {
		logstore = logstore[:len(logstore)-1]
	}

	// 创建运行时选项
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 查询SLS索引信息
	middleware.Logger.Infof("开始查询SLS索引信息: 项目=%s, 日志库=%s", project, logstore)
	response, err := client.GetIndexWithOptions(
		tea.String(project),
		tea.String(logstore),
		headers,
		runtime,
	)

	if err != nil {
		// 处理错误
		var errMsg string
		if sdkErr, ok := err.(*tea.SDKError); ok {
			errMsg = *sdkErr.Message
		} else {
			errMsg = err.Error()
		}
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("查询SLS索引失败: %v", errMsg))
		return
	}

	// 构建字段映射
	fields := make(map[string]interface{})
	properties := make(map[string]interface{})

	// 处理索引配置
	if response != nil && response.Body != nil && response.Body.Keys != nil {
		for key, attr := range response.Body.Keys {
			fieldType := "text" // 默认为文本类型

			// 根据SLS类型转换为ES类型
			if attr.Type != nil {
				switch *attr.Type {
				case "long":
					fieldType = "long"
				case "double":
					fieldType = "double"
				case "json":
					fieldType = "object"
				}
			}

			// 添加到属性中
			properties[key] = map[string]interface{}{
				"type": fieldType,
			}
		}
	}

	// 将字段映射添加到结果中
	fields["properties"] = properties

	// 获取日志库中存在的索引名称（在SLS中，索引名称通常与日志库名称相同）
	indices := []string{logstore}

	// 返回组合结果
	result := map[string]interface{}{
		"indices": indices, // 所有匹配的索引列表
		"fields":  fields,  // 字段信息
	}

	middleware.Logger.Infof("SLS索引查询成功，找到 %d 个字段", len(properties))
	middleware.ResponseSuccess(w, "获取SLS索引信息成功", result)
}
