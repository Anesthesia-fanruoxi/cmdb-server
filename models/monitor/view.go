package monitor

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"fmt"
	"time"
)

// MonitorView 监控视图模型
type MonitorView struct {
	ID          int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ViewName    string    `json:"view_name" gorm:"column:view_name;type:varchar(100);not null"`
	Query       string    `json:"query" gorm:"column:query;type:text;not null"`
	Description string    `json:"description" gorm:"column:description;type:varchar(255)"`
	ChartType   string    `json:"chart_type" gorm:"column:chart_type;type:varchar(50)"`
	CreatedBy   int64     `json:"created_by" gorm:"column:created_by;not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;type:datetime(3)"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;type:datetime(3)"`
}

// TableName 设置表名
func (MonitorView) TableName() string {
	return middleware.TableMonitorViews
}

// MonitorViewList 视图列表带分页
type MonitorViewList struct {
	Total int64         `json:"total"`
	Items []MonitorView `json:"items"`
}

// CreateView 创建监控视图
func CreateView(_ context.Context, view *MonitorView) error {
	// 使用GORM创建记录
	result := database.DB.Create(view)
	return result.Error
}

// GetViewByID 根据ID获取视图
func GetViewByID(_ context.Context, id int64) (*MonitorView, error) {
	var view MonitorView

	// 使用GORM查询单条记录
	result := database.DB.First(&view, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &view, nil
}

// UpdateView 更新监控视图
func UpdateView(_ context.Context, view *MonitorView) error {
	// 使用GORM更新记录
	result := database.DB.Model(&MonitorView{}).Where("id = ?", view.ID).Updates(map[string]interface{}{
		"view_name":   view.ViewName,
		"query":       view.Query,
		"description": view.Description,
		"chart_type":  view.ChartType,
	})

	return result.Error
}

// DeleteView 删除监控视图
func DeleteView(_ context.Context, id int64) error {
	// 使用GORM删除记录
	result := database.DB.Delete(&MonitorView{}, id)
	return result.Error
}

// ListViews 分页获取视图列表
func ListViews(_ context.Context, offset, limit int, category string) (*MonitorViewList, error) {
	var total int64
	var items []MonitorView

	// 构建查询
	query := database.DB.Model(&MonitorView{})

	// 添加过滤条件
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 查询总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %v", err)
	}

	// 查询数据
	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("查询数据失败: %v", err)
	}

	return &MonitorViewList{
		Total: total,
		Items: items,
	}, nil
}
