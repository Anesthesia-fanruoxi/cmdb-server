package export

import (
	"cmdb/models/sql"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/pkg/notice"
	"cmdb/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SqlExportCreate 创建SQL审批的API处理函数
func SqlExportCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 检查HTTP方法
	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "读取请求数据失败")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	// 从上下文获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "无法获取用户信息")
		return
	}

	// 解析请求
	var req sql.CreateExportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "解析请求数据失败")
		return
	}

	if req.SubmitterRemark == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少申请说明")
		return
	}

	// 如果提供了数据库名称，则SQL内容也必须提供
	if req.DatabaseName != "" && req.SQLContent == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "提供数据库名称时必须同时提供SQL内容")
		return
	}

	// 直接查询用户名信息
	var applyName string
	var reviewerName string
	var executorName string

	// 固定审核人ID为60（硬编码）
	const fixedReviewerID int64 = 60

	// 查询审批人姓名（从请求参数获取）
	result := database.DB.Table(middleware.TableUsers).
		Select("user_name").
		Where("id = ?", req.ApplyID).
		Limit(1).
		Row()
	if err := result.Scan(&applyName); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "未找到审批人信息")
		return
	}

	// 查询审核人姓名（固定为UserID=60）
	result = database.DB.Table(middleware.TableUsers).
		Select("user_name").
		Where("id = ?", fixedReviewerID).
		Limit(1).
		Row()
	if err := result.Scan(&reviewerName); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "未找到审核人信息（UserID=60）")
		return
	}

	middleware.Logger.Infof("使用固定审核人: ID=%d, Name=%s", fixedReviewerID, reviewerName)

	// 查询执行人姓名
	result = database.DB.Table(middleware.TableUsers).
		Select("user_name").
		Where("id = ?", req.ExecutorID).
		Limit(1).
		Row()
	if err := result.Scan(&executorName); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "未找到执行人信息")
		return
	}

	// 生成随机URL用于审批链接
	exportUrl := sql.GenerateRandomURL()
	middleware.Logger.Infof("已生成审批随机URL: %s", exportUrl)

	// 获取当前中国时区时间
	chinaTime := middleware.GetChinaTime()

	// 创建审批记录
	export := &sql.SQLExport{
		Project:         fmt.Sprintf("%d", req.Project), // 项目ID转字符串存储
		SubmitterID:     userInfo.UserID,
		SubmitterName:   userInfo.UserName,
		SubmitterRemark: req.SubmitterRemark,
		SubmitterStatus: 1,           // 提交后不可修改
		ApplyID:         req.ApplyID, // 审批人ID从请求参数获取
		ApplyName:       applyName,
		ApplyStatus:     1,               // 审批人可以处理
		ReviewerID:      fixedReviewerID, // 固定审核人ID=60
		ReviewerName:    reviewerName,
		ReviewerStatus:  0, // 还不能处理
		ExecutorID:      req.ExecutorID,
		ExecutorName:    executorName,
		ExecutorStatus:  0, // 还不能处理
		Status:          0, // 0-待提交
		CreatedAt:       chinaTime,
		UpdatedAt:       chinaTime,
		// ApplyTime 和 ReviewerTime 应该在实际审批/审核时才设置，不在创建时设置
		UrlKey:       exportUrl,
		DatabaseName: req.DatabaseName, // 可选的数据库名称
		SQLContent:   req.SQLContent,   // 可选的SQL内容
	}

	// 保存到数据库
	if err := database.DB.Create(export).Error; err != nil {
		middleware.ResponseError(w, http.StatusInternalServerError, "创建审批记录失败: "+err.Error())
		return
	}

	// 记录已保存到数据库
	middleware.Logger.Infof("审批记录已保存到数据库，ID: %d", export.ID)

	// 保存审批信息到Redis
	if err := sql.SaveExportToRedis(export, exportUrl); err != nil {
		middleware.Logger.Errorf("保存审批信息到Redis失败: %v", err)
	}

	// 获取系统设置
	var setting system.Setting
	if err := database.DB.First(&setting).Error; err != nil {
		middleware.Logger.Errorf("获取系统设置失败: %v", err)
	} else {
		// 拼接审批URL
		ApplyUrl := fmt.Sprintf("%s/export?id=%s", setting.SystemUrl, exportUrl)
		TextMsg := "新的SQL导出请求，请点击链接查看：" + ApplyUrl
		// 发送飞书群组通知
		if err := notice.SendFeiShuText(TextMsg); err != nil {
			middleware.Logger.Errorf("发送飞书群组通知失败: %v", err)
		} else {
			middleware.Logger.Infof("飞书群组通知已发送，URL: %s", ApplyUrl)
		}
	}

	// 创建成功后，发送卡片给审批人
	go func() {
		if err := sql.SendApprovalNotificationCard(export, "approver"); err != nil {
			middleware.Logger.Errorf("发送审批人卡片通知失败: %v", err)
		} else {
			middleware.Logger.Infof("已发送审批卡片通知给审批人: %s", export.ApplyName)
		}
	}()

	// 返回成功结果
	middleware.ResponseSuccess(w, "创建审批记录成功", export)
}
