package system

import (
	"cmdb/initData"
	"cmdb/models"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

// GetDeptProjects 获取部门关联的项目列表
func GetDeptProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 获取部门ID
	deptIDStr := r.URL.Query().Get("dept_id")
	if deptIDStr == "" {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "部门ID不能为空")
		return
	}

	deptID, err := strconv.ParseUint(deptIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的部门ID")
		return
	}

	var deptProjects []models.DeptProject
	if err := initData.GetDB().Where("dept_id = ?", deptID).Find(&deptProjects).Error; err != nil {
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "获取部门项目关联失败")
		return
	}

	// 提取项目标识列表
	projects := make([]string, len(deptProjects))
	for i, dp := range deptProjects {
		projects[i] = dp.Project
	}

	utils.Success(w, projects)
}

// UpdateDeptProjects 更新部门关联的项目列表
func UpdateDeptProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, utils.ERROR, "方法不允许")
		return
	}

	// 解析请求体
	var req struct {
		ID       uint     `json:"id"`       // 部门ID
		Projects []string `json:"projects"` // 项目标识列表
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "无效的请求参数")
		return
	}

	if req.ID == 0 {
		utils.Error(w, http.StatusBadRequest, utils.INVALID_PARAMS, "部门ID不能为空")
		return
	}

	// 检查部门是否存在
	var dept models.Department
	if err := initData.GetDB().First(&dept, req.ID).Error; err != nil {
		utils.Error(w, http.StatusNotFound, utils.NOT_FOUND, "部门不存在")
		return
	}

	db := initData.GetDB()
	tx := db.Begin()

	// 删除原有关联
	if err := tx.Delete(&models.DeptProject{}, "dept_id = ?", req.ID).Error; err != nil {
		tx.Rollback()
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "删除原有关联失败")
		return
	}

	// 添加新的关联
	if len(req.Projects) > 0 {
		deptProjects := make([]models.DeptProject, len(req.Projects))
		for i, project := range req.Projects {
			deptProjects[i] = models.DeptProject{
				DeptID:  req.ID,
				Project: project,
			}
		}

		if err := tx.Create(&deptProjects).Error; err != nil {
			tx.Rollback()
			utils.Error(w, http.StatusInternalServerError, utils.ERROR, "创建新关联失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.Error(w, http.StatusInternalServerError, utils.ERROR, "更新关联关系失败")
		return
	}

	utils.Success(w, nil)
}
