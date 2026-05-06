package export

import (
	"cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

// SqlExportDetail 导出申请详情处理函数
func SqlExportDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 检查HTTP方法
	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 从上下文中获取用户信息
	userInfo, ok := r.Context().Value(utils.UserInfoKey).(utils.UserInfo)
	if !ok {
		middleware.ResponseError(w, http.StatusUnauthorized, "无法获取用户信息")
		return
	}

	// 获取URL参数中的ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少必要的ID参数")
		return
	}

	// 转换ID为整数
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "无效的ID参数")
		return
	}

	// 构建查询 - 使用别名避免列名冲突
	db := database.DB.Table(middleware.TableSQLExport + " AS se")

	// 连接项目字典表获取项目中文名称
	db = db.Joins("LEFT JOIN " + sql.TableProjectDict + " AS pd ON se.project = pd.id")

	// 连接用户表获取用户详细信息
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS submitter ON se.submitter_id = submitter.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS applier ON se.apply_id = applier.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS reviewer ON se.reviewer_id = reviewer.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS executor ON se.executor_id = executor.id")

	// 选择需要的字段
	db = db.Select(`
		se.id, 
		COALESCE(pd.project, se.project) AS project,
		COALESCE(pd.project_name, se.project) AS project_name,
		COALESCE(submitter.nick_name, se.submitter_name) AS submitter_name,
		se.submitter_id,
		se.submitter_status,
		se.submitter_remark,
		COALESCE(applier.nick_name, se.apply_name) AS apply_name,
		se.apply_id,
		se.apply_status,
		se.sql_content,
		se.apply_source,
		se.apply_time,
		se.database_name,
		COALESCE(reviewer.nick_name, se.reviewer_name) AS reviewer_name,
		se.reviewer_id,
		se.reviewer_status,
		se.reviewer_time,
		se.reviewer_source,
		COALESCE(executor.nick_name, se.executor_name) AS executor_name,
		se.executor_id,
		se.executor_status,
		se.executor_source,
		se.created_at,
		se.status
	`)

	// 根据ID查询记录
	db = db.Where("se.id = ?", id)

	// 获取记录
	var result sql.ExportQueryResult
	if err := db.First(&result).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "未找到指定的导出申请记录")
		return
	}

	// 检查用户是否有权限查看该记录
	// 简单权限验证：记录的提交人、审批人或执行人才能查看
	if userInfo.UserID != 1 && // 假设ID为1的用户是管理员，总是有权限
		result.SubmitterID != userInfo.UserID &&
		result.ApplyID != userInfo.UserID &&
		result.ReviewerID != userInfo.UserID &&
		result.ExecutorID != userInfo.UserID {
		middleware.ResponseError(w, http.StatusForbidden, "您没有权限查看此记录")
		return
	}

	// 转换为输出格式
	detail := sql.ConvertQueryResultToDetail(&result)

	// 返回成功结果
	middleware.ResponseSuccess(w, "获取导出申请详情成功", detail)
}

// SqlExportUrlDetail 通过URL获取SQL导出详情的API处理函数
func SqlExportUrlDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持GET请求")
		return
	}

	// 获取URL参数中的ID
	urlID := r.URL.Query().Get("id")
	if urlID == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "缺少必要的ID参数")
		return
	}
	middleware.Logger.Infof("接受到请求，获取urlID:%s", urlID)
	// 从Redis获取导出信息
	redisKey := middleware.GetSqlExportKey(urlID)
	middleware.Logger.Infof("Redis完整的键值：%s", redisKey)
	exportJSON, err := database.RDB.Get(r.Context(), redisKey).Result()
	if err != nil {
		middleware.Logger.Errorf("Redis获取导出信息失败 [url=%s]: %v", urlID, err)
		middleware.ResponseError(w, http.StatusNotFound, "无效的导出链接")
		return
	}

	// 解析导出信息
	var export sql.SQLExport
	if err := json.Unmarshal([]byte(exportJSON), &export); err != nil {
		middleware.Logger.Errorf("解析导出信息失败: %v", err)
		middleware.ResponseError(w, http.StatusInternalServerError, "解析导出信息失败")
		return
	}

	// 查询导出记录以获取最新状态
	db := database.DB.Table(middleware.TableSQLExport + " AS se")

	// 连接项目字典表获取项目中文名称
	db = db.Joins("LEFT JOIN " + sql.TableProjectDict + " AS pd ON se.project = pd.id")

	// 连接用户表获取用户详细信息
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS submitter ON se.submitter_id = submitter.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS applier ON se.apply_id = applier.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS reviewer ON se.reviewer_id = reviewer.id")
	db = db.Joins("LEFT JOIN " + sql.TableUsers + " AS executor ON se.executor_id = executor.id")

	// 选择需要的字段
	db = db.Select(`
		se.id, 
		COALESCE(pd.project, se.project) AS project,
		COALESCE(pd.project_name, se.project) AS project_name,
		COALESCE(submitter.nick_name, se.submitter_name) AS submitter_name,
		se.submitter_id,
		se.submitter_status,
		se.submitter_remark,
		COALESCE(applier.nick_name, se.apply_name) AS apply_name,
		se.apply_id,
		se.apply_status,
		se.sql_content,
		se.apply_source,
		se.apply_time,
		se.database_name,
		COALESCE(reviewer.nick_name, se.reviewer_name) AS reviewer_name,
		se.reviewer_id,
		se.reviewer_status,
		se.reviewer_time,
		se.reviewer_source,
		COALESCE(executor.nick_name, se.executor_name) AS executor_name,
		se.executor_id,
		se.executor_status,
		se.executor_source,
		se.created_at,
		se.status
	`)

	// 根据ID查询记录
	db = db.Where("se.id = ?", export.ID)

	// 获取记录
	var result sql.ExportQueryResult
	if err := db.First(&result).Error; err != nil {
		middleware.ResponseError(w, http.StatusNotFound, "未找到指定的导出申请记录")
		return
	}

	// 转换为输出格式
	detail := sql.ConvertQueryResultToDetail(&result)

	// 返回成功结果
	middleware.ResponseSuccess(w, "获取导出申请详情成功", detail)
}
