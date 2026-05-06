package view

import (
	"cmdb/api/elfk/search"
	"cmdb/models/elfk"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ViewExpireTime 视图信息在Redis中的过期时间
const ViewExpireTime = 24 * time.Hour

// InitializeViewDetailInfo 使用直接请求代理的方式初始化视图详情
// 与GetViewDetailInfo不同，这个方法直接请求ES代理获取信息，不依赖HTTP接口
// 专门用于系统初始化阶段
func InitializeViewDetailInfo(_ context.Context, view elfk.View) (map[string]interface{}, error) {
	// 获取索引字段信息和索引列表（直接使用 project 名称）
	fields, indices, err := search.GetEsFieldsDirectly(view.Project, view.IndexPattern)
	if err != nil {
		return nil, fmt.Errorf("直接获取ES字段信息失败: %v", err)
	}

	// 构建视图响应结构
	viewResponse := elfk.ViewResponse{
		ID:           view.ID,
		Name:         view.Name,
		Project:      view.Project,
		Category:     view.Category,
		LogType:      view.LogType,
		IndexPattern: view.IndexPattern,
		TimeField:    view.TimeField,
		TimeFormat:   view.TimeFormat,
		Description:  view.Description,
		UpdateTime:   view.UpdateTime,
	}

	// 构建完整的视图信息
	return buildViewResponse(viewResponse, fields, indices), nil
}

// saveViewToRedis 将视图信息保存到Redis
func SaveViewToRedis(ctx context.Context, viewID uint, viewInfo map[string]interface{}) error {
	// 将视图信息序列化为JSON
	viewJSON, err := json.Marshal(viewInfo)
	if err != nil {
		return fmt.Errorf("序列化视图信息失败: %v", err)
	}

	// 获取Redis键
	redisKey := middleware.GetESViewKey(viewID)

	// 存储到Redis
	return database.RDB.Set(ctx, redisKey, viewJSON, ViewExpireTime).Err()
}

// GetViewFromRedis 从Redis获取视图信息
func GetViewFromRedis(ctx context.Context, viewID uint) (map[string]interface{}, error) {
	// 获取Redis键
	redisKey := middleware.GetESViewKey(viewID)

	// 从Redis获取视图信息
	viewJSON, err := database.RDB.Get(ctx, redisKey).Bytes()
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var viewInfo map[string]interface{}
	if err := json.Unmarshal(viewJSON, &viewInfo); err != nil {
		return nil, fmt.Errorf("解析视图信息失败: %v", err)
	}

	return viewInfo, nil
}

// UpdateViewRedisCache 更新单个视图的Redis缓存
// 当视图被创建或更新时调用此函数，确保Redis缓存与数据库同步
// 只有类型为elfk的视图才会被缓存到Redis
func UpdateViewRedisCache(viewID uint) error {
	ctx := context.Background()

	// 从数据库中获取视图信息
	var view elfk.View
	if err := database.DB.First(&view, viewID).Error; err != nil {
		middleware.Logger.Errorf("更新视图缓存失败：数据库查询错误 [ID:%d] - %v", viewID, err)
		return err
	}

	// 只处理elfk类型的视图
	if view.LogType != "elfk" {
		middleware.Logger.Infof("视图跳过缓存 [ID:%d, 项目:%s, 名称:%s, 类型:%s] - 非elfk类型",
			viewID, view.Project, view.Name, view.LogType)
		return nil
	}

	// 获取视图的详细信息
	viewInfo, err := InitializeViewDetailInfo(ctx, view)
	if err != nil {
		middleware.Logger.Errorf("更新视图缓存失败：获取详情错误 [ID:%d, 项目:%s, 名称:%s] - %v",
			viewID, view.Project, view.Name, err)
		return err
	}

	// 将视图信息存入Redis
	if err := SaveViewToRedis(ctx, viewID, viewInfo); err != nil {
		middleware.Logger.Errorf("更新视图缓存失败：存入Redis错误 [ID:%d, 项目:%s, 名称:%s] - %v",
			viewID, view.Project, view.Name, err)
		return err
	}

	middleware.Logger.Infof("视图缓存已更新 [ID:%d, 项目:%s, 名称:%s]",
		viewID, view.Project, view.Name)
	return nil
}
