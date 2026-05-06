package task

import (
	"bytes"
	"fmt"
	"text/template"
)

// RenderScript 渲染脚本模板
// templateContent: 脚本模板内容，使用 {{变量名}} 作为占位符
// params: 参数映射，key为变量名，value为实际值
// 返回渲染后的脚本内容
func RenderScript(templateContent string, params map[string]interface{}) (string, error) {
	// 如果参数为空，直接返回原内容
	if len(params) == 0 {
		return templateContent, nil
	}

	// 创建模板
	tmpl, err := template.New("script").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("模板解析失败: %v", err)
	}

	// 执行模板渲染
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("模板渲染失败: %v", err)
	}

	return buf.String(), nil
}
