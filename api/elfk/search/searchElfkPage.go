package search

import (
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"fmt"
)

// HandleESPagedSearch 处理ES分页查询
// 从Redis缓存中获取数据并分页
func HandleESPagedSearch(ctx context.Context, pageReq ESSearchPageRequest) (*ESSearchResult, error) {
	// 验证参数
	if pageReq.QueryID == "" {
		return nil, fmt.Errorf("查询ID不能为空")
	}

	// 设置默认分页参数
	if pageReq.Page <= 0 {
		pageReq.Page = 1
	}

	// 使用默认页大小
	pageSize := DefaultPageSize

	// 记录日志
	middleware.Logger.Infof("处理ELFK缓存分页: QID=%s, Page=%d, Size=%d",
		pageReq.QueryID, pageReq.Page, pageSize)

	// 从缓存获取分页结果
	pagedResult, err := GetESPagedResult(ctx, database.RDB, pageReq.QueryID, pageReq.Page, pageSize)
	if err != nil {
		middleware.Logger.Errorf("获取ELFK分页结果失败: %v", err)
		return nil, err
	}

	return pagedResult, nil
}
