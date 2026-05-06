package menu

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	ID         int64  `json:"id" validate:"required"`                 // 菜单ID，必填
	ParentID   int64  `json:"parent_id" validate:"omitempty,gte=0"`   // 父菜单ID，选填
	Name       string `json:"name" validate:"omitempty,min=2,max=32"` // 菜单名称，选填
	Path       string `json:"path" validate:"omitempty,max=128"`      // 路由路径，选填
	Component  string `json:"component" validate:"omitempty,max=128"` // 组件路径，选填
	Sort       int    `json:"sort" validate:"omitempty,gte=0"`        // 排序，选填
	Icon       string `json:"icon" validate:"omitempty,max=32"`       // 图标，选填
	Permission string `json:"permission" validate:"omitempty,max=32"` //权限标识，选填
	IsEnabled  bool   `json:"is_enabled"`                             // 是否启用，选填
}

// MenuUpdate 更新菜单
func MenuUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 读取请求体
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求体失败")
		return
	}

	// 解析为原始 map 以检查字段是否存在
	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 解析为请求结构体
	var req UpdateMenuRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 查询菜单是否存在
	var menu system.Menu
	if err := database.DB.First(&menu, req.ID).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "菜单不存在")
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	// 只更新提供的字段
	if _, exists := rawData["parent_id"]; exists {
		// 检查父菜单是否存在
		if req.ParentID > 0 {
			var parentMenu system.Menu
			if err := database.DB.First(&parentMenu, req.ParentID).Error; err != nil {
				middleware.ResponseError(w, http.StatusBadRequest, "父菜单不存在")
				return
			}
		}
		updates["parent_id"] = req.ParentID
	}

	if _, exists := rawData["name"]; exists {
		updates["name"] = req.Name
	}

	if _, exists := rawData["path"]; exists {
		updates["path"] = req.Path
	}

	if _, exists := rawData["component"]; exists {
		updates["component"] = req.Component
	}

	if _, exists := rawData["sort"]; exists {
		updates["sort"] = req.Sort
	}

	if _, exists := rawData["icon"]; exists {
		updates["icon"] = req.Icon
	}

	if _, exists := rawData["is_enabled"]; exists {
		updates["is_enabled"] = req.IsEnabled
	}

	if _, exists := rawData["permission"]; exists {
		updates["permission"] = req.Permission
	}

	// 如果没有要更新的字段
	if len(updates) == 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "没有要更新的字段")
		return
	}

	// 执行更新
	if err := database.DB.Model(&menu).Updates(updates).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "更新菜单失败")
		return
	}

	middleware.ResponseSuccess(w, "更新成功", nil)
}
