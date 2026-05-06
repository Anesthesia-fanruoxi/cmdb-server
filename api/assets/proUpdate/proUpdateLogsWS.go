package proUpdate

import (
	"cmdb/config"
	"cmdb/models/assets"
	"cmdb/models/system"
	"cmdb/pkg/middleware"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// 日志代理管理器
type LogProxyManager struct {
	proxies map[string]*LogProxy
	mutex   sync.RWMutex
}

// 全局代理管理器
var proxyManager = &LogProxyManager{
	proxies: make(map[string]*LogProxy),
}

// 日志代理结构
type LogProxy struct {
	taskID       string
	stepType     string
	project      string
	taskType     string
	remote       string
	frontendConn *websocket.Conn
	backendConn  *websocket.Conn
	backendURL   string
	done         chan bool
	mutex        sync.Mutex
}

// ProUpdateLogsWS WebSocket日志代理
// GET /api/assets/proUpdate/logs/ws?task_id={task_id}&step_type={step_type}&project_name={project_name}&type={type}&remote={remote}
func ProUpdateLogsWS(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	taskID := r.URL.Query().Get("task_id")
	stepType := r.URL.Query().Get("step_type")
	projectName := r.URL.Query().Get("project")
	taskType := r.URL.Query().Get("type")
	remote := r.URL.Query().Get("remote")

	middleware.Logger.Infof("收到WebSocket连接请求 - TaskID: %s, StepType: %s, Project: %s, Type: %s, Remote: %s, RemoteAddr: %s",
		taskID, stepType, projectName, taskType, remote, r.RemoteAddr)

	if taskID == "" || stepType == "" || projectName == "" {
		middleware.Logger.Errorf("WebSocket连接参数不完整 - TaskID: %s, StepType: %s, Project: %s", taskID, stepType, projectName)
		middleware.ResponseError(w, http.StatusBadRequest, "缺少必要参数task_id、step_type或project_name")
		return
	}

	// 从 Redis 中查询对应步骤的 remote 值（忽略URL传入的remote参数）
	ctx := context.Background()
	remoteFromRedis, err := getRemoteFromRedis(ctx, taskID, stepType)
	if err != nil {
		middleware.Logger.Errorf("从Redis获取remote失败 - TaskID: %s, StepType: %s, Error: %v", taskID, stepType, err)
		middleware.ResponseError(w, http.StatusInternalServerError, fmt.Sprintf("获取任务信息失败: %v", err))
		return
	}
	remote = remoteFromRedis
	middleware.Logger.Infof("从Redis获取到remote值: %s - TaskID: %s, StepType: %s", remote, taskID, stepType)

	// 校验 remote 值
	if remote != "server" && remote != "agent" {
		middleware.ResponseError(w, http.StatusBadRequest, "remote 仅支持 server 或 agent")
		return
	}

	// 升级为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		middleware.Logger.Errorf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	middleware.Logger.Infof("WebSocket连接升级成功 - TaskID: %s, StepType: %s", taskID, stepType)

	// 创建日志代理
	proxy := &LogProxy{
		taskID:       taskID,
		stepType:     stepType,
		project:      projectName,
		taskType:     taskType,
		remote:       remote,
		frontendConn: conn,
		done:         make(chan bool),
	}

	// 启动代理
	middleware.Logger.Infof("开始启动日志代理 - TaskID: %s, StepType: %s", taskID, stepType)
	if err := proxy.Start(); err != nil {
		middleware.Logger.Errorf("启动日志代理失败: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("连接失败: %v", err)))
		return
	}
	middleware.Logger.Infof("日志代理启动成功 - TaskID: %s, StepType: %s", taskID, stepType)

	// 等待连接结束
	<-proxy.done
	middleware.Logger.Infof("日志代理连接结束, TaskID: %s, StepType: %s", taskID, stepType)
}

// getRemoteFromRedis 从Redis中获取指定任务和步骤的remote值
func getRemoteFromRedis(ctx context.Context, taskID, stepType string) (string, error) {
	// 获取任务记录
	recordJSON, err := GetTaskRecord(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("获取任务记录失败: %v", err)
	}

	// 解析任务记录
	var taskRecords []assets.RedisTaskRecord
	if err := json.Unmarshal([]byte(recordJSON), &taskRecords); err != nil {
		return "", fmt.Errorf("解析任务记录失败: %v", err)
	}

	if len(taskRecords) == 0 {
		return "", fmt.Errorf("未找到任务记录")
	}

	taskRecord := taskRecords[0]

	// 查找对应步骤的remote值
	for _, step := range taskRecord.StepDetail {
		if step.StepType == stepType {
			if step.Remote != "" {
				return step.Remote, nil
			}
		}
	}

	// 如果没有找到对应步骤或步骤中没有remote值，默认返回agent
	return "agent", nil
}

// 启动代理
func (p *LogProxy) Start() error {
	// 1. 构建后端服务器WebSocket URL
	// 从配置文件读取服务地址
	cfg := config.GetConfig()

	// 根据 remote 选择目标地址和加密密钥
	var baseURL string
	var routeBy string
	var salt string

	if p.remote == "server" {
		// 直接走 cicd-service
		baseURL = cfg.CICD.URL
		salt = cfg.CICD.Secret
		routeBy = "server"
		middleware.Logger.Infof("WebSocket连接server - URL: %s, Salt长度: %d", baseURL, len(salt))
	} else {
		// agent: 提取项目标识（处理 -web 后缀）
		projectIdentifier := extractProjectIdentifier(p.project)
		middleware.Logger.Infof("WebSocket日志 - 项目名称: %s, 提取标识: %s", p.project, projectIdentifier)

		// 从数据库获取项目的 agent_url
		var err error
		baseURL, err = system.GetProjectAgentURL(projectIdentifier)
		if err != nil {
			return fmt.Errorf("获取项目 %s 的Agent地址失败: %v", projectIdentifier, err)
		}
		salt = cfg.Agent.Secret
		routeBy = fmt.Sprintf("agent(project=%s)", projectIdentifier)
		middleware.Logger.Infof("项目 %s 使用Agent地址: %s", projectIdentifier, baseURL)
	}

	// 检查加密密钥
	if salt == "" {
		return fmt.Errorf("系统配置错误：缺少加密密钥 (remote=%s)", p.remote)
	}

	// 将HTTP地址转换为WebSocket地址
	wsURL := baseURL
	if len(wsURL) > 7 && wsURL[:7] == "http://" {
		wsURL = "ws://" + wsURL[7:]
	} else if len(wsURL) > 8 && wsURL[:8] == "https://" {
		wsURL = "wss://" + wsURL[8:]
	}

	// 构建参数结构
	params := map[string]string{
		"taskId":   p.taskID,
		"project":  p.project,
		"stepType": p.stepType,
		"type":     p.taskType,
	}

	// 序列化参数
	paramsData, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("序列化WebSocket参数失败: %v", err)
	}

	middleware.Logger.Infof("WebSocket参数(加密前) - %s", string(paramsData))

	// 加密参数
	encryptedParams, err := middleware.CompressAndEncrypt(paramsData, salt)
	if err != nil {
		return fmt.Errorf("加密WebSocket参数失败: %v", err)
	}

	// URL编码加密参数
	encodedParams := url.QueryEscape(encryptedParams)

	p.backendURL = fmt.Sprintf("%s/ws/cicd/logs?data=%s", wsURL, encodedParams)

	middleware.Logger.Infof("准备连接后端服务器WebSocket - RouteBy: %s, URL: %s", routeBy, p.backendURL)

	// 2. 连接B服务器WebSocket
	backendConn, _, err := websocket.DefaultDialer.Dial(p.backendURL, nil)
	if err != nil {
		return fmt.Errorf("连接B服务器WebSocket失败: %v", err)
	}
	p.backendConn = backendConn
	middleware.Logger.Infof("成功连接到B服务器WebSocket - TaskID: %s, StepType: %s", p.taskID, p.stepType)

	// 3. 注册到管理器
	proxyKey := fmt.Sprintf("%s:%s", p.taskID, p.stepType)
	proxyManager.mutex.Lock()
	proxyManager.proxies[proxyKey] = p
	proxyManager.mutex.Unlock()
	middleware.Logger.Infof("代理已注册到管理器 - Key: %s", proxyKey)

	// 4. 启动消息转发
	middleware.Logger.Infof("启动消息转发协程 - TaskID: %s, StepType: %s", p.taskID, p.stepType)
	go p.forwardFromBackend()
	go p.handleFrontendClose()

	return nil
}

// 从B服务器转发消息到前端
func (p *LogProxy) forwardFromBackend() {
	defer p.cleanup()

	middleware.Logger.Infof("开始从B服务器转发消息 - TaskID: %s, StepType: %s", p.taskID, p.stepType)

	for {
		messageType, message, err := p.backendConn.ReadMessage()
		if err != nil {
			// 检查是否是正常关闭连接
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) ||
				strings.Contains(err.Error(), "closed") {
				middleware.Logger.Debugf("后端连接已关闭 - TaskID: %s, StepType: %s", p.taskID, p.stepType)
			} else {
				middleware.Logger.Errorf("从B服务器读取消息失败 - TaskID: %s, StepType: %s, Error: %v", p.taskID, p.stepType, err)
			}
			break
		}

		middleware.Logger.Debugf("收到B服务器消息 - TaskID: %s, StepType: %s, MessageType: %d, Length: %d",
			p.taskID, p.stepType, messageType, len(message))

		p.mutex.Lock()
		if p.frontendConn != nil {
			err = p.frontendConn.WriteMessage(messageType, message)
			if err != nil {
				middleware.Logger.Errorf("向前端发送消息失败 - TaskID: %s, StepType: %s, Error: %v", p.taskID, p.stepType, err)
				p.mutex.Unlock()
				break
			}
			middleware.Logger.Debugf("成功转发消息到前端 - TaskID: %s, StepType: %s", p.taskID, p.stepType)
		} else {
			middleware.Logger.Warnf("前端连接已断开，无法转发消息 - TaskID: %s, StepType: %s", p.taskID, p.stepType)
		}
		p.mutex.Unlock()
	}

	middleware.Logger.Infof("停止从B服务器转发消息 - TaskID: %s, StepType: %s", p.taskID, p.stepType)
}

// 处理前端连接关闭
func (p *LogProxy) handleFrontendClose() {
	defer p.cleanup()

	middleware.Logger.Infof("开始监听前端连接状态 - TaskID: %s, StepType: %s", p.taskID, p.stepType)

	for {
		_, _, err := p.frontendConn.ReadMessage()
		if err != nil {
			middleware.Logger.Infof("前端连接关闭 - TaskID: %s, StepType: %s, Error: %v", p.taskID, p.stepType, err)
			break
		}
	}
}

// 清理资源
func (p *LogProxy) cleanup() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 关闭后端连接
	if p.backendConn != nil {
		p.backendConn.Close()
		p.backendConn = nil
	}

	// 关闭前端连接
	if p.frontendConn != nil {
		p.frontendConn.Close()
		p.frontendConn = nil
	}

	// 从管理器移除
	proxyKey := fmt.Sprintf("%s:%s", p.taskID, p.stepType)
	proxyManager.mutex.Lock()
	delete(proxyManager.proxies, proxyKey)
	proxyManager.mutex.Unlock()

	// 发送完成信号
	select {
	case p.done <- true:
	default:
	}
}
