package initialize

import (
	elfkview "cmdb/api/elfk/view"
	"cmdb/models/elfk"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"context"
	"sync"
	"time"
)

// InitViewsRedisCache 初始化视图Redis缓存
// 在系统启动时调用此函数，将所有类型为elfk的视图信息缓存到Redis
func InitViewsRedisCache() {
	// 开始记录处理时间
	startTime := time.Now()

	// 查询数据库获取所有elfk类型的视图
	var views []elfk.View
	if err := database.DB.Where("log_type = ?", "elfk").Find(&views).Error; err != nil {
		middleware.Logger.Errorf("初始化视图缓存失败，获取视图列表错误: %v", err)
		return
	}

	if len(views) == 0 {
		middleware.Logger.Info("初始化视图缓存：没有找到elfk类型的视图")
		return
	}

	// 使用WaitGroup来等待所有异步处理完成
	var wg sync.WaitGroup
	// 记录成功和失败的数量
	successCount := 0
	failCount := 0

	// 创建一个并发控制通道，限制最大并发数为10
	concurrencyLimit := make(chan struct{}, 10)

	// 创建一个互斥锁，用于保护计数器
	var mutex sync.Mutex

	ctx := context.Background()

	// 对每个视图异步获取详细信息并存入Redis
	for _, view := range views {
		wg.Add(1)

		// 将视图对象拷贝到协程中，避免闭包陷阱
		currentView := view

		go func() {
			defer wg.Done()

			// 获取并发槽
			concurrencyLimit <- struct{}{}
			// 函数结束时释放并发槽
			defer func() { <-concurrencyLimit }()

			// 使用新的直接方法获取视图的详细信息，而不是通过HTTP接口
			viewInfo, err := elfkview.InitializeViewDetailInfo(ctx, currentView)

			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				middleware.Logger.Errorf("初始化缓存：获取视图详情失败 [ID:%d, 项目:%s, 名称:%s] - %v",
					currentView.ID, currentView.Project, currentView.Name, err)
				failCount++
			} else {
				// 将视图信息存入Redis
				if err := elfkview.SaveViewToRedis(ctx, currentView.ID, viewInfo); err != nil {
					middleware.Logger.Errorf("初始化缓存：视图信息存入Redis失败 [ID:%d, 项目:%s, 名称:%s] - %v",
						currentView.ID, currentView.Project, currentView.Name, err)
					failCount++
				} else {
					middleware.Logger.Infof("初始化缓存：视图缓存成功 [ID:%d, 项目:%s, 名称:%s]",
						currentView.ID, currentView.Project, currentView.Name)
					successCount++
				}
			}
		}()
	}

	// 等待所有处理完成
	wg.Wait()

	// 计算总耗时
	elapsedTime := time.Since(startTime).Seconds()

	middleware.Logger.Infof("初始化elfk视图缓存完成，总数: %d，成功: %d，失败: %d，耗时: %.2f秒",
		len(views), successCount, failCount, elapsedTime)
}

// 定时刷新视图缓存的时间点
var ViewCacheRefreshHour = 1 // 默认凌晨1点刷新

// StartViewCacheRefreshScheduler 启动定时刷新视图缓存的调度器
// 在系统启动时调用此函数，每天在指定时间自动更新视图缓存
func StartViewCacheRefreshScheduler() {
	middleware.Logger.Info("启动视图缓存定时刷新调度器")

	// 在后台协程中运行定时刷新
	go func() {
		// 延迟一定时间后再执行第一次初始化，避免与系统启动冲突
		middleware.Logger.Info("系统启动完成后将进行视图缓存初始化，等待1秒...")
		time.Sleep(1 * time.Second)

		// 执行第一次初始化
		middleware.Logger.Info("开始执行视图缓存初始化")
		InitViewsRedisCache()

		for {
			// 计算距离下次执行的时间
			nextRun := getNextRunTime(ViewCacheRefreshHour)
			middleware.Logger.Infof("下次视图缓存刷新时间: %v", nextRun.Format("2006-01-02 15:04:05"))

			// 等待到达指定时间
			timer := time.NewTimer(time.Until(nextRun))
			<-timer.C

			// 执行缓存刷新
			middleware.Logger.Info("定时刷新视图缓存开始")
			InitViewsRedisCache()
		}
	}()

	middleware.Logger.Infof("视图缓存定时刷新调度器已在后台启动，将在每天 %d:00 刷新", ViewCacheRefreshHour)
}

// getNextRunTime 计算下一次执行的时间点
func getNextRunTime(hour int) time.Time {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())

	// 如果今天的执行时间已经过了，计算明天的执行时间
	if now.After(today) {
		return today.Add(24 * time.Hour)
	}

	// 否则返回今天的执行时间
	return today
}
