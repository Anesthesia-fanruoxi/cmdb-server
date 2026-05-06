package user

import (
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
)

type CreateUserRequest struct {
	UserName string `json:"user_name" label:"用户名" validate:"required,min=3,max=32"`
	Password string `json:"password" label:"密码"`
	NickName string `json:"nick_name" label:"昵称" validate:"omitempty,max=32"`
	Email    string `json:"email" label:"邮箱" validate:"omitempty,email,max=128"`
	Phone    string `json:"phone" label:"手机号" validate:"omitempty,len=11"`
	RoleID   int64  `json:"role_id" label:"角色ID" validate:"omitempty"`
	DeptID   int64  `json:"dept_id" label:"部门ID" validate:"omitempty"`
}

// UserCreate 创建用户
func UserCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从上下文获取当前用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "获取用户信息失败")
		return
	}

	// 获取当前用户的角色级别
	var currentRole system.Role
	if err := database.DB.Table(middleware.TableRoles).
		Select("level").
		Where("id = ?", userInfo.RoleID).
		First(&currentRole).Error; err != nil {
		middleware.ResponseError(w, http.StatusForbidden, "无效的角色")
		return
	}

	// 解析请求体
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	// 验证请求参数
	if err := middleware.Validator.Struct(req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, middleware.TranslateError(err))
		return
	}

	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&system.User{}).Where("user_name = ?", req.UserName).Count(&count)
	if count > 0 {
		middleware.ResponseError(w, http.StatusBadRequest, "用户名已存在")
		return
	}

	// 获取系统设置中的默认密码
	var setting system.Setting
	if err := database.DB.Table(middleware.TableSettings).First(&setting).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "获取系统设置失败")
		return
	}

	// 处理密码
	password := req.Password
	if password == "" {
		password = setting.DefaultPassword
	} else {
		// 如果提供了密码，验证密码复杂度
		if err := utils.ValidatePassword(password); err != nil {
			middleware.ResponseError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// 加密密码
	hashedPassword, err := utils.EncryptPassword(password)
	if err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "密码加密失败")
		return
	}

	// 根据角色级别处理创建逻辑
	var roleID, deptID uint
	if req.RoleID == 1 || req.DeptID == 1 {
		middleware.ResponseError(w, http.StatusBadRequest, "不允许创建系统角色或者系统部门用户")
	}

	if currentRole.Level == 0 { // 超级管理员
		// 处理角色ID
		if req.RoleID > 0 {
			roleID = middleware.FromInt64[uint](req.RoleID)
			// 检查指定的角色是否存在
			var role system.Role
			if err := database.DB.First(&role, roleID).Error; err != nil {
				middleware.ResponseError(w, http.StatusBadRequest, "指定的角色不存在")
				return
			}
		} else {
			roleID = 4 // 默认为普通用户角色
		}

		// 处理部门ID
		if req.DeptID > 0 {
			deptID = middleware.FromInt64[uint](req.DeptID)
			// 检查指定的部门是否存在
			var dept system.Department
			if err := database.DB.First(&dept, deptID).Error; err != nil {
				middleware.ResponseError(w, http.StatusBadRequest, "指定的部门不存在")
				return
			}
		} else {
			deptID = 99 // 默认部门ID
		}
	} else { // 普通管理员
		// 只能创建普通用户角色，部门与自己相同
		roleID = 4 // 普通用户角色
		if req.DeptID > 0 {
			// 检查是否是自己部门的子部门
			var dept system.Department
			if err := database.DB.First(&dept, req.DeptID).Error; err != nil {
				middleware.ResponseError(w, http.StatusBadRequest, "指定的部门不存在")
				return
			}
			// TODO: 这里可以添加检查是否是子部门的逻辑
			deptID = middleware.FromInt64[uint](req.DeptID)
		} else {
			deptID = middleware.FromInt64[uint](userInfo.DeptID)
		}
	}

	// 创建用户
	user := system.User{
		UserName:  req.UserName,
		Password:  hashedPassword,
		NickName:  req.NickName,
		Email:     req.Email,
		Phone:     req.Phone,
		RoleID:    int64(roleID),
		DeptID:    int64(deptID),
		IsEnabled: true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建用户失败")
		return
	}

	// 返回成功响应
	middleware.ResponseSuccess(w, "用户创建成功", map[string]interface{}{
		"user_id": user.ID,
	})
}
