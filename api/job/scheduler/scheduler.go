package scheduler

import (
	"bytes"
	"cmdb/config"
	"cmdb/models/job"
	"cmdb/models/system"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// JobScheduler 任务调度器
type JobScheduler struct {
	cron      *cron.Cron
	jobs      map[int64]cron.EntryID // job_id -> cron_entry_id
	mu        sync.RWMutex
	agentURLs map[string]string // project -> agent_url
}

var (
	globalScheduler *JobScheduler
	once            sync.Once
)

// GetScheduler 获取全局调度器实例
func GetScheduler() *JobScheduler {
	once.Do(func() {
		// 加载中国时区
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			middleware.Logger.Errorf("加载中国时区失败，使用固定东八区: %v", err)
			loc = time.FixedZone("CST", 8*3600) // 使用固定的东八区
		}

		globalScheduler = &JobScheduler{
			cron:      cron.New(cron.WithSeconds(), cron.WithLocation(loc)),
			jobs:      make(map[int64]cron.EntryID),
			agentURLs: make(map[string]string),
		}
		middleware.Logger.Infof("任务调度器初始化完成，时区: %s", loc.String())
	})
	return globalScheduler
}

// Start 启动调度器
func (s *JobScheduler) Start() {
	s.cron.Start()
	middleware.Logger.Info("任务调度器已启动")

	// 加载所有启用的任务
	s.LoadJobs()
}

// Stop 停止调度器
func (s *JobScheduler) Stop() {
	s.cron.Stop()
	middleware.Logger.Info("任务调度器已停止")
}

// LoadJobs 从数据库加载所有启用的任务
func (s *JobScheduler) LoadJobs() error {
	var jobs []job.JobInfo
	if err := database.DB.Where("status = ?", 1).Find(&jobs).Error; err != nil {
		return fmt.Errorf("加载任务失败: %v", err)
	}

	for i := range jobs {
		// 使用索引访问，避免循环变量引用问题
		if err := s.AddJob(&jobs[i]); err != nil {
			middleware.Logger.Errorf("添加任务失败 [JobID: %d, TaskKey: %s]: %v", jobs[i].ID, jobs[i].TaskKey, err)
		} else {
			middleware.Logger.Infof("任务已添加 [JobID: %d, TaskKey: %s]", jobs[i].ID, jobs[i].TaskKey)
		}
	}

	return nil
}

// AddJob 添加任务到调度器
func (s *JobScheduler) AddJob(jobInfo *job.JobInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果任务已存在，先移除
	if entryID, exists := s.jobs[jobInfo.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, jobInfo.ID)
	}

	// 只添加启用状态的任务
	if jobInfo.Status != 1 {
		return nil
	}

	// 查询该任务关联的所有项目
	var projects []job.JobProject
	if err := database.DB.Where("job_id = ?", jobInfo.ID).Find(&projects).Error; err != nil {
		return fmt.Errorf("查询任务关联项目失败: %v", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("任务未关联任何项目")
	}

	// 创建局部副本，避免闭包引用问题
	currentJob := *jobInfo
	currentProjects := make([]job.JobProject, len(projects))
	copy(currentProjects, projects)

	// 添加定时任务
	entryID, err := s.cron.AddFunc(currentJob.Cron, func() {
		s.executeJob(&currentJob, currentProjects)
	})

	if err != nil {
		return fmt.Errorf("添加cron任务失败: %v", err)
	}

	s.jobs[jobInfo.ID] = entryID
	return nil
}

// RemoveJob 从调度器移除任务
func (s *JobScheduler) RemoveJob(jobID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.jobs[jobID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, jobID)
		middleware.Logger.Infof("任务已移除 [JobID: %d]", jobID)
	}
}

// UpdateJob 更新任务
func (s *JobScheduler) UpdateJob(jobInfo *job.JobInfo) error {
	// 先移除旧任务
	s.RemoveJob(jobInfo.ID)

	// 重新添加任务
	if err := s.AddJob(jobInfo); err != nil {
		return fmt.Errorf("更新任务失败: %v", err)
	}

	middleware.Logger.Infof("任务已更新 [JobID: %d, TaskKey: %s]", jobInfo.ID, jobInfo.TaskKey)
	return nil
}

// executeJob 执行任务 - 推送给Agent
func (s *JobScheduler) executeJob(jobInfo *job.JobInfo, projects []job.JobProject) {
	middleware.Logger.Infof("开始执行任务 [JobID: %d, TaskKey: %s]", jobInfo.ID, jobInfo.TaskKey)

	// 获取系统设置中的回调地址
	var setting system.Setting
	if err := database.DB.Table(middleware.TableSettings).First(&setting).Error; err != nil {
		middleware.Logger.Errorf("获取系统设置失败: %v", err)
		return
	}

	callbackURL := setting.SystemUrl
	if callbackURL == "" {
		middleware.Logger.Error("系统回调地址未配置")
		return
	}
	middleware.Logger.Infof("系统回调地址: %s", callbackURL)

	// 获取配置
	cfg := config.GetConfig()

	// 为每个项目推送任务
	for _, proj := range projects {
		// 获取Agent URL
		agentURL, err := s.getAgentURL(proj.Project)
		if err != nil {
			middleware.Logger.Errorf("获取Agent地址失败 [Project: %s]: %v", proj.Project, err)
			continue
		}
		middleware.Logger.Infof("获取到Agent地址 [Project: %s, AgentURL: %s]", proj.Project, agentURL)

		// 记录开始时间
		startTime := time.Now()

		// 创建执行记录（状态：执行中）
		execution := &job.JobExecution{
			JobID:      jobInfo.ID,
			Project:    proj.Project,
			TaskKey:    jobInfo.TaskKey,
			TaskName:   jobInfo.Name,
			ExecStatus: 0, // 0:执行中
			StartTime:  startTime,
			CreatedAt:  startTime,
		}

		if err := database.DB.Create(execution).Error; err != nil {
			middleware.Logger.Errorf("创建执行记录失败 [JobID: %d, Project: %s]: %v", jobInfo.ID, proj.Project, err)
			continue
		}

		middleware.Logger.Infof("创建执行记录 [ExecutionID: %d, JobID: %d, Project: %s]", execution.ID, jobInfo.ID, proj.Project)

		// 构建执行任务请求（只发送脚本ID和参数）
		executeReq := map[string]interface{}{
			"job_id":       execution.ID, // 执行记录ID
			"task_name":    jobInfo.Name,
			"task_key":     jobInfo.TaskKey,
			"callback_url": callbackURL + "/api/job/agent/report",
		}
		middleware.Logger.Infof("任务请求 [JobID: %d, TaskKey: %s]",
			jobInfo.ID, jobInfo.TaskKey)

		// 序列化请求数据
		reqJSON, err := json.Marshal(executeReq)
		if err != nil {
			middleware.Logger.Errorf("序列化请求失败: %v", err)
			continue
		}
		middleware.Logger.Infof("序列化后的JSON数据: %s", string(reqJSON))

		// 加密请求数据
		encryptedData, err := middleware.CompressAndEncrypt(reqJSON, cfg.Agent.Secret)
		if err != nil {
			middleware.Logger.Errorf("加密请求失败: %v", err)
			continue
		}
		middleware.Logger.Infof("加密后数据长度: %d", len(encryptedData))

		// 发送加密数据
		reqBody := map[string]string{
			"data": encryptedData,
		}
		reqBodyJSON, _ := json.Marshal(reqBody)

		// 构建完整的Agent URL
		fullAgentURL := agentURL + "/proxy/cron-plugs/api/task/execute"
		middleware.Logger.Infof("准备推送任务 [JobID: %d, Project: %s, URL: %s]", jobInfo.ID, proj.Project, fullAgentURL)

		resp, err := http.Post(fullAgentURL, "application/json", bytes.NewBuffer(reqBodyJSON))
		if err != nil {
			middleware.Logger.Errorf("推送任务失败 [Project: %s, AgentURL: %s]: %v", proj.Project, fullAgentURL, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// 读取响应body
			respBody, _ := io.ReadAll(resp.Body)
			middleware.Logger.Errorf("Agent响应异常 [Project: %s, URL: %s, Status: %d, Response: %s]", proj.Project, fullAgentURL, resp.StatusCode, string(respBody))
			continue
		}

		// 读取成功响应
		respBody, _ := io.ReadAll(resp.Body)
		middleware.Logger.Infof("任务推送成功 [JobID: %d, Project: %s, Response: %s]", jobInfo.ID, proj.Project, string(respBody))
	}
}

// getAgentURL 获取项目对应的Agent地址
func (s *JobScheduler) getAgentURL(project string) (string, error) {
	// 先从缓存查询
	s.mu.RLock()
	url, exists := s.agentURLs[project]
	s.mu.RUnlock()

	if exists && url != "" {
		return url, nil
	}

	// 从数据库查询项目字典表
	var projectDict system.ProjectDict
	if err := database.DB.Where("project = ?", project).First(&projectDict).Error; err != nil {
		return "", fmt.Errorf("未找到项目信息: %v", err)
	}

	// 检查 agent_url 是否配置
	if projectDict.AgentURL == "" {
		return "", fmt.Errorf("项目未配置Agent地址")
	}

	// 缓存起来
	s.mu.Lock()
	s.agentURLs[project] = projectDict.AgentURL
	s.mu.Unlock()

	return projectDict.AgentURL, nil
}

// SetAgentURL 手动设置项目的Agent地址
func (s *JobScheduler) SetAgentURL(project, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agentURLs[project] = url
}
