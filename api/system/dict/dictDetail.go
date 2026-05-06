package dict

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"fmt"
	"net/http"
	"strings"
)

// DictDetail 获取字典详情
func DictDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 获取字典ID
	tableName := r.URL.Query().Get("target_table")
	if tableName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "表名不能为空")
		return
	}

	// 获取字典记录
	var dictRecord system.DictRecord
	if err := database.DB.Where("target_table = ?", tableName).First(&dictRecord).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "字典不存在")
		return
	}

	// 如果是项目字典，需要验证权限
	if tableName == middleware.TableProjectDict {
		// 构建项目查询条件
		query := fmt.Sprintf(`
			SELECT 
				id,
				t.%s as `+"`key`"+`, 
				t.%s as `+"`value`"+`, 
				t.color as color 
			FROM %s t`,
			dictRecord.KeyName,
			dictRecord.ValueName,
			tableName,
		)

		// 获取项目权限
		projects, isSuperAdmin, err := utils.CheckRequestProjectPermission(r, "prod")
		if err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// 如果不是超级管理员，只查询有权限的项目
		if !isSuperAdmin {
			if len(projects) > 0 {
				query += fmt.Sprintf(" WHERE t.%s IN ('%s')", dictRecord.KeyName, strings.Join(projects, "','"))
			} else {
				// 如果没有可访问的项目，返回空列表
				middleware.ResponseSuccess(w, "获取字典详情成功", map[string]interface{}{
					"dict":  dictRecord,
					"items": []map[string]interface{}{},
				})
				return
			}
		}

		var items []map[string]interface{}
		if err := database.DB.Raw(query).Scan(&items).Error; err != nil {
			middleware.ResponseError(w, http.StatusInternalServerError, "获取字典项失败")
			return
		}

		middleware.ResponseSuccess(w, "获取字典详情成功", map[string]interface{}{
			"dict":  dictRecord,
			"items": items,
		})
		return
	}

	// 非项目字典，直接查询所有数据
	query := fmt.Sprintf(`
		SELECT 
			id,
			t.%s as `+"`key`"+`, 
			t.%s as `+"`value`"+`, 
			t.color as color 
		FROM %s t`,
		dictRecord.KeyName,
		dictRecord.ValueName,
		tableName,
	)

	var items []map[string]interface{}
	if err := database.DB.Raw(query).Scan(&items).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取字典项失败")
		return
	}

	middleware.ResponseSuccess(w, "获取字典详情成功", map[string]interface{}{
		"dict":  dictRecord,
		"items": items,
	})
}
