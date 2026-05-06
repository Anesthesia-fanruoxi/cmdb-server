package sql

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"cmdb/pkg/notice"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// SQLExport 导出流程表
type SQLExport struct {
	ID              int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement;comment:主键ID"`
	Project         string     `json:"project" gorm:"column:project;not null;comment:项目名称"`
	SubmitterID     int64      `json:"submitterId" gorm:"column:submitter_id;not null;comment:提交人ID"`
	SubmitterName   string     `json:"submitterName" gorm:"column:submitter_name;not null;comment:提交人姓名"`
	SubmitterStatus int        `json:"submitterStatus" gorm:"column:submitter_status;default:0;comment:提交人状态: 0-不能处理，1-可以处理"`
	SubmitterRemark string     `json:"submitterRemark" gorm:"column:submitter_remark;comment:申请说明"`
	ApplyID         int64      `json:"applyId" gorm:"column:apply_id;comment:审批人ID"`
	ApplyName       string     `json:"applyName" gorm:"column:apply_name;comment:审批人姓名"`
	ApplyTime       *time.Time `json:"applyTime" gorm:"column:apply_time;comment:审批时间"`
	ApplySource     string     `json:"applySource" gorm:"column:apply_source;comment:审批来源"`
	ApplyStatus     int        `json:"applyStatus" gorm:"column:apply_status;default:0;comment:审批人状态：0-不能处理，1-可以处理"`
	SQLContent      string     `json:"sqlContent" gorm:"column:sql_content;comment:SQL内容"`
	DatabaseName    string     `json:"databaseName" gorm:"column:database_name;comment:目标数据库名称"`
	ReviewerID      int64      `json:"reviewerId" gorm:"column:reviewer_id;comment:审核人ID"`
	ReviewerName    string     `json:"reviewerName" gorm:"column:reviewer_name;comment:审核人姓名"`
	ReviewerTime    *time.Time `json:"reviewerTime" gorm:"column:reviewer_time;comment:审核时间"`
	ReviewerSource  string     `json:"reviewerSource" gorm:"column:reviewer_source;comment:审核来源"`
	ReviewerStatus  int        `json:"reviewerStatus" gorm:"column:reviewer_status;default:0;comment:审核人状态：0-不能处理，1-可以处理"`
	ExecutorID      int64      `json:"executorId" gorm:"column:executor_id;comment:执行人ID"`
	ExecutorName    string     `json:"executorName" gorm:"column:executor_name;comment:执行人姓名"`
	ExecutorSource  string     `json:"executorSource" gorm:"column:executor_source;comment:执行人来源"`
	ExecutionResult string     `json:"executionResult" gorm:"column:execution_result;comment:执行结果"`
	UrlKey          string     `json:"urlKey" gorm:"column:url_key;comment:URL键"`
	ExecutorStatus  int        `json:"executorStatus" gorm:"column:executor_status;default:0;comment:执行人状态：0-不能处理，1-可以处理"`
	// 文件相关字段
	FilePath      string `json:"filePath" gorm:"column:file_path;comment:文件存储路径"`
	FileName      string `json:"fileName" gorm:"column:file_name;comment:文件名"`
	FileSize      int64  `json:"fileSize" gorm:"column:file_size;default:0;comment:文件大小(字节)"`
	DownloadCount int    `json:"downloadCount" gorm:"column:download_count;default:0;comment:下载次数"`
	DownloadLimit int    `json:"downloadLimit" gorm:"column:download_limit;default:3;comment:下载次数限制"`
	// 时间和状态字段
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
	Status    int       `json:"status" gorm:"column:status;default:0;comment:状态：0-待提交，1-待审核，2-待执行，3-执行中，4-执行完成，5-执行失败，6-已驳回，7-已撤销"`
}

// TableName 指定表名
func (SQLExport) TableName() string {
	return middleware.TableSQLExport
}

// 导出流程状态常量
const (
	StatusDraft     = 0 // 待提交
	StatusAudit     = 1 // 待审核
	StatusExec      = 2 // 待执行
	StatusExecuting = 3 // 执行中
	StatusCompleted = 4 // 执行成功
	StatusCanceled  = 5 // 已撤销
	StatusRejected  = 6 // 已拒绝
	StatusFailed    = 7 // 执行失败
	StatusExporting = 8 // 导出中
)

// 角色状态常量
const (
	StatusCannotProcess = 0 // 不能处理
	StatusCanProcess    = 1 // 可以处理
)

// Redis 相关常量
const (
	// SQLExportExpireTime SQL审批信息在Redis中的过期时间
	SQLExportExpireTime = 24 * time.Hour
)

// 表名常量
const (
	TableUsers       = "sys_users"        // 用户表
	TableProjectDict = "sys_project_dict" // 项目字典表
)

// 初始化随机数生成器
var randSource = rand.NewSource(time.Now().UnixNano())
var randGenerator = rand.New(randSource)

// 用于生成随机URL的字符集
const urlCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 状态映射
var StatusTextMap = map[int]string{
	StatusDraft:     "待提交",
	StatusAudit:     "待审核",
	StatusExec:      "待执行",
	StatusExecuting: "执行中",
	StatusCompleted: "已完成",
	StatusCanceled:  "已撤销",
	StatusRejected:  "已拒绝",
	StatusFailed:    "执行失败",
	StatusExporting: "导出中",
}

// CreateExportRequest 创建SQL审批请求的参数
type CreateExportRequest struct {
	Project         int    `json:"project" binding:"required"`           // 项目ID
	ApplyID         int64  `json:"apply_id" binding:"required"`          // 审批人ID
	ReviewerID      int64  `json:"reviewer_id" binding:"required"`       // 审核人ID
	ExecutorID      int64  `json:"executor_id" binding:"required"`       // 执行人ID
	SubmitterRemark string `json:"submitter_remark"  binding:"required"` // 申请说明
	DatabaseName    string `json:"database_name,omitempty"`              // 数据库名称（可选）
	SQLContent      string `json:"sql_content,omitempty"`                // SQL内容（可选）
}

// UpdateExportRequest 更新导出流程请求
type UpdateExportRequest struct {
	ID           int64  `json:"id"`                      // 流程ID
	ProcessType  int    `json:"process_type"`            // 处理类型：1通过，0拒绝
	DatabaseName string `json:"database_name,omitempty"` // 数据库名称（仅审批人需要）
	SQLContent   string `json:"sql_content,omitempty"`   // SQL内容（仅审批人需要）
}

// SqlExportDetailRe 导出申请详情结构体
type SqlExportDetailRe struct {
	ID              int64  `json:"id"`               // 主键ID
	Project         string `json:"project"`          // 项目名称（简称）
	ProjectName     string `json:"project_name"`     // 项目名称（中文）
	SubmitterName   string `json:"submitter_name"`   // 提交人姓名
	SubmitterID     int64  `json:"submitter_id"`     // 提交人ID
	SubmitterStatus int    `json:"submitter_status"` // 提交人状态: 0-不能处理，1-可以处理
	SubmitterRemark string `json:"submitter_remark"` // 申请说明
	ApplyName       string `json:"apply_name"`       // 审批人姓名
	ApplyID         int64  `json:"apply_id"`         // 审批人ID
	ApplyStatus     int    `json:"apply_status"`     // 审批人状态: 0-不能处理，1-可以处理
	SQLContent      string `json:"sql_content"`      // SQL内容
	ApplySource     string `json:"apply_source"`     // 审批来源
	ApplyTime       string `json:"apply_time"`       // 审批时间
	DatabaseName    string `json:"database_name"`    // 目标数据库名称
	ReviewerName    string `json:"reviewer_name"`    // 审核人姓名
	ReviewerID      int64  `json:"reviewer_id"`      // 审核人ID
	ReviewerStatus  int    `json:"reviewer_status"`  // 审核人状态: 0-不能处理，1-可以处理
	ReviewerTime    string `json:"reviewer_time"`    // 审核时间
	ReviewerSource  string `json:"reviewer_source"`  // 审核来源
	ExecutorName    string `json:"executor_name"`    // 执行人姓名
	ExecutorID      int64  `json:"executor_id"`      // 执行人ID
	ExecutorStatus  int    `json:"executor_status"`  // 执行人状态: 0-不能处理，1-可以处理
	ExecutorSource  string `json:"executor_source"`  // 执行人来源
	CreatedAt       string `json:"created_at"`       // 创建时间
	Status          int    `json:"status"`           // 状态
	StatusText      string `json:"status_text"`      // 状态描述
	CurrentOperator string `json:"current_operator"` // 当前操作人
}

// UrlDetailRequest URL详情请求参数
type UrlDetailRequest struct {
	Url      string `json:"url" binding:"required"`       // 链接
	UserName string `json:"user_name" binding:"required"` // 用户名
	Code     string `json:"code" binding:"required"`      // 验证码
}

// SqlExportItem 精简后的SQL审批项
type SqlExportItem struct {
	ID              int64  `json:"id"`
	Project         string `json:"project"`
	ProjectName     string `json:"project_name"`
	SubmitterName   string `json:"submitter_name"`   // 提交人
	ApplyName       string `json:"apply_name"`       // 审批人
	ReviewerName    string `json:"reviewer_name"`    // 审核人
	ExecutorName    string `json:"executor_name"`    // 执行人
	CreatedAt       string `json:"created_at"`       // 创建时间
	StatusText      string `json:"status_text"`      // 当前状态（中文）
	Status          int    `json:"status"`           // 当前状态（数字）
	CurrentOperator string `json:"current_operator"` // 当前操作人
}

// EmailInfo 邮件接收者信息
type EmailInfo struct {
	To      []string // 主收件人列表
	Cc      []string // 抄送人列表
	Subject string   // 可选，自定义邮件主题
}

// SQLExecuteResult SQL执行结果
type SQLExecuteResult struct {
	SuccessCount  int      `json:"success_count"`   // 成功执行的SQL数量
	FailedCount   int      `json:"failed_count"`    // 执行失败的SQL数量
	AffectedRows  int64    `json:"affected_rows"`   // 受影响的行数（DML语句）
	ResultCount   int64    `json:"result_count"`    // 查询结果行数（SELECT语句）
	Errors        []string `json:"errors"`          // 错误信息列表
	ExcelFilePath string   `json:"excel_file_path"` // 生成的Excel文件路径
	ExecutionTime int64    `json:"execution_time"`  // 执行时间(毫秒)
	EmailSent     bool     `json:"email_sent"`      // 邮件是否发送成功
}

// ExportQueryResult 查询结果结构体 - 用于查询数据库时的结构映射
type ExportQueryResult struct {
	ID              int64     `gorm:"column:id"`
	Project         string    `gorm:"column:project"`
	ProjectName     string    `gorm:"column:project_name"`
	SubmitterName   string    `gorm:"column:submitter_name"`
	SubmitterID     int64     `gorm:"column:submitter_id"`
	SubmitterStatus int       `gorm:"column:submitter_status"`
	SubmitterRemark string    `gorm:"column:submitter_remark"`
	ApplyName       string    `gorm:"column:apply_name"`
	ApplyID         int64     `gorm:"column:apply_id"`
	ApplyStatus     int       `gorm:"column:apply_status"`
	SQLContent      string    `gorm:"column:sql_content"`
	ApplySource     string    `gorm:"column:apply_source"`
	ApplyTime       time.Time `gorm:"column:apply_time"`
	DatabaseName    string    `gorm:"column:database_name"`
	ReviewerName    string    `gorm:"column:reviewer_name"`
	ReviewerID      int64     `gorm:"column:reviewer_id"`
	ReviewerStatus  int       `gorm:"column:reviewer_status"`
	ReviewerTime    time.Time `gorm:"column:reviewer_time"`
	ReviewerSource  string    `gorm:"column:reviewer_source"`
	ExecutorName    string    `gorm:"column:executor_name"`
	ExecutorID      int64     `gorm:"column:executor_id"`
	ExecutorStatus  int       `gorm:"column:executor_status"`
	ExecutorSource  string    `gorm:"column:executor_source"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	Status          int       `gorm:"column:status"`
}

// GenerateRandomURL 生成随机的8位大小写混合字符串
func GenerateRandomURL() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = urlCharset[randGenerator.Intn(len(urlCharset))]
	}
	return string(b)
}

// GetStatusText 获取状态的中文描述
func GetStatusText(status int) string {
	text, ok := StatusTextMap[status]
	if !ok {
		return "未知状态"
	}
	return text
}

// FormatTimeField 格式化时间字段
func FormatTimeField(t time.Time) string {
	return middleware.FormatTimeToChina(t)
}

// FormatTimeFieldPtr 格式化时间指针字段
func FormatTimeFieldPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return middleware.FormatTimeToChina(*t)
}

// GetUserEmail 根据用户ID获取用户邮箱
func GetUserEmail(userID int64) string {
	// 如果用户ID为0，返回空
	if userID <= 0 {
		return ""
	}

	var email string
	// 查询用户邮箱
	err := database.DB.Table(middleware.TableUsers).
		Select("email").
		Where("id = ?", userID).
		Row().
		Scan(&email)

	if err != nil {
		middleware.Logger.Errorf("获取用户 %d 邮箱失败: %v", userID, err)
		return ""
	}

	middleware.Logger.Infof("获取用户 %d 邮箱: %s", userID, email)
	return email
}

// GetProjectName 根据项目标识获取项目中文名称
func GetProjectName(projectID string) string {
	// 如果项目ID为空，返回空
	if projectID == "" {
		return ""
	}

	var projectName string
	// 查询项目中文名称
	err := database.DB.Table(TableProjectDict).
		Select("project_name").
		Where("project = ?", projectID).
		Row().
		Scan(&projectName)

	if err != nil {
		middleware.Logger.Errorf("获取项目 %s 中文名称失败: %v", projectID, err)
		return ""
	}

	middleware.Logger.Infof("获取项目 %s 中文名称: %s", projectID, projectName)
	return projectName
}

// SaveExportToRedis 保存审批信息到Redis
func SaveExportToRedis(export *SQLExport, exportUrl string) error {
	ctx := context.Background()

	// 将审批信息序列化为JSON
	exportJSON, err := json.Marshal(export)
	if err != nil {
		return fmt.Errorf("序列化审批信息失败: %v", err)
	}

	// 保存完整的审批信息
	redisKey := middleware.GetSqlExportKey(exportUrl)
	if err := database.RDB.Set(ctx, redisKey, string(exportJSON), SQLExportExpireTime).Err(); err != nil {
		return fmt.Errorf("保存审批信息到Redis失败: %v", err)
	}

	middleware.Logger.Infof("审批信息已保存到Redis，Key: %s", redisKey)
	return nil
}

// UpdateExportRedis 更新Redis中的导出记录
func UpdateExportRedis(export *SQLExport) error {
	ctx := context.Background()
	if export.UrlKey == "" {
		return fmt.Errorf("UrlKey 为空，无法定位 Redis key")
	}
	key := middleware.GetSqlExportKey(export.UrlKey)

	// 序列化并更新Redis
	exportJSON, err := json.Marshal(export)
	if err != nil {
		return fmt.Errorf("序列化导出信息失败: %v", err)
	}

	if err := database.RDB.Set(ctx, key, string(exportJSON), SQLExportExpireTime).Err(); err != nil {
		return fmt.Errorf("更新Redis失败: %v", err)
	}

	return nil
}

// CalculateCurrentOperator 计算当前操作人
func CalculateCurrentOperator(export *SQLExport) string {
	switch export.Status {
	case StatusDraft: // 待提交
		if export.ApplyStatus == StatusCanProcess {
			return export.ApplyName
		}
		return export.SubmitterName
	case StatusAudit: // 待审核
		return export.ReviewerName
	case StatusExec: // 待执行
		return export.ExecutorName
	case StatusExecuting, StatusExporting: // 执行中/导出中
		return export.ExecutorName
	case StatusFailed: // 执行失败（执行人可重新导出）
		return export.ExecutorName
	}
	return "" // 其他状态没有当前操作人
}

// CalculateCurrentOperatorFromResult 从查询结果计算当前操作人
func CalculateCurrentOperatorFromResult(result *ExportQueryResult) string {
	switch result.Status {
	case StatusDraft: // 待提交
		if result.ApplyStatus == StatusCanProcess {
			return result.ApplyName
		}
		return result.SubmitterName
	case StatusAudit: // 待审核
		return result.ReviewerName
	case StatusExec: // 待执行
		return result.ExecutorName
	case StatusExecuting, StatusExporting: // 执行中/导出中
		return result.ExecutorName
	case StatusFailed: // 执行失败（执行人可重新导出）
		return result.ExecutorName
	}
	return "" // 其他状态没有当前操作人
}

// ConvertQueryResultToDetail 将查询结果转换为详情响应
func ConvertQueryResultToDetail(result *ExportQueryResult) SqlExportDetailRe {
	return SqlExportDetailRe{
		ID:              result.ID,
		Project:         result.Project,
		ProjectName:     result.ProjectName,
		SubmitterName:   result.SubmitterName,
		SubmitterID:     result.SubmitterID,
		SubmitterStatus: result.SubmitterStatus,
		SubmitterRemark: result.SubmitterRemark,
		ApplyName:       result.ApplyName,
		ApplyID:         result.ApplyID,
		ApplyStatus:     result.ApplyStatus,
		SQLContent:      result.SQLContent,
		ApplySource:     result.ApplySource,
		ApplyTime:       FormatTimeField(result.ApplyTime),
		DatabaseName:    result.DatabaseName,
		ReviewerName:    result.ReviewerName,
		ReviewerID:      result.ReviewerID,
		ReviewerStatus:  result.ReviewerStatus,
		ReviewerTime:    FormatTimeField(result.ReviewerTime),
		ReviewerSource:  result.ReviewerSource,
		ExecutorName:    result.ExecutorName,
		ExecutorID:      result.ExecutorID,
		ExecutorStatus:  result.ExecutorStatus,
		ExecutorSource:  result.ExecutorSource,
		CreatedAt:       FormatTimeField(result.CreatedAt),
		Status:          result.Status,
		StatusText:      GetStatusText(result.Status),
		CurrentOperator: CalculateCurrentOperatorFromResult(result),
	}
}

// ExportListResult 列表查询结果结构体 - 用于列表查询时的结构映射
type ExportListResult struct {
	ID            int64     `gorm:"column:id"`
	Project       string    `gorm:"column:project"`
	ProjectName   string    `gorm:"column:project_name"`
	SubmitterName string    `gorm:"column:submitter_name"`
	ApplyName     string    `gorm:"column:apply_name"`
	ApplyStatus   int       `gorm:"column:apply_status"`
	ReviewerName  string    `gorm:"column:reviewer_name"`
	ExecutorName  string    `gorm:"column:executor_name"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	Status        int       `gorm:"column:status"`
}

// CalculateCurrentOperatorFromListResult 从列表查询结果计算当前操作人
func CalculateCurrentOperatorFromListResult(result *ExportListResult) string {
	switch result.Status {
	case StatusDraft: // 待提交
		if result.ApplyStatus == StatusCanProcess {
			return result.ApplyName
		}
		return result.SubmitterName
	case StatusAudit: // 待审核
		return result.ReviewerName
	case StatusExec: // 待执行
		return result.ExecutorName
	case StatusExecuting, StatusExporting: // 执行中/导出中
		return result.ExecutorName
	case StatusFailed: // 执行失败（执行人可重新导出）
		return result.ExecutorName
	}
	return "" // 其他状态没有当前操作人
}

// SendApprovalNotificationCard 发送审批卡片通知
func SendApprovalNotificationCard(export *SQLExport, role string) error {
	var targetUserID int64
	var targetEmail string

	// 根据角色确定目标用户
	switch role {
	case "approver":
		targetUserID = export.ApplyID
	case "reviewer":
		targetUserID = export.ReviewerID
	case "executor":
		targetUserID = export.ExecutorID
	default:
		return fmt.Errorf("不支持的角色: %s", role)
	}

	// 获取目标用户邮箱
	targetEmail = GetUserEmail(targetUserID)
	if targetEmail == "" {
		return fmt.Errorf("无法获取用户 %d 的邮箱", targetUserID)
	}

	// 获取项目中文名称
	projectName := GetProjectName(export.Project)
	if projectName == "" {
		projectName = export.Project // 如果获取不到中文名，使用原项目标识
	}

	// 检查是否已有SQL
	hasSQL := export.SQLContent != "" && export.DatabaseName != ""

	// 构建卡片信息
	cardInfo := notice.ApprovalCardInfo{
		ExportID:        export.ID,
		SubmitterName:   export.SubmitterName,
		ProjectName:     projectName,
		DatabaseName:    export.DatabaseName,
		SubmitterRemark: export.SubmitterRemark,
		SQLContent:      export.SQLContent,
		CreatedAt:       FormatTimeField(export.CreatedAt),
		Role:            role,
		HasSQL:          hasSQL,
	}

	// 发送卡片
	return notice.SendApprovalActionCard(targetEmail, cardInfo)
}
