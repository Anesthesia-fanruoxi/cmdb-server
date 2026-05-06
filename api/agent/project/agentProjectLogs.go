package project

import (
	"cmdb/models/system"
	"cmdb/pkg/middleware"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// AgentProjectLogs WebSocket代理，实时查看插件日志
func AgentProjectLogs(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	project := r.URL.Query().Get("project")
	pluginName := r.URL.Query().Get("name")

	if project == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "project参数不能为空")
		return
	}
	if pluginName == "" {
		middleware.ResponseError(w, http.StatusBadRequest, "name参数不能为空")
		return
	}

	// 获取项目Agent URL
	agentURL, err := system.GetProjectAgentURL(project)
	if err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, fmt.Sprintf("获取项目Agent地址失败: %v", err))
		return
	}

	// 将http(s)://转换为ws(s)://
	wsURL := strings.Replace(agentURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	// 构建Agent WebSocket URL
	agentWSURL := fmt.Sprintf("%s/ws/plugins/logs?name=%s", wsURL, pluginName)

	// 升级客户端连接为WebSocket
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer clientConn.Close()

	// 连接到Agent的WebSocket
	agentConn, resp, err := websocket.DefaultDialer.Dial(agentWSURL, nil)
	if err != nil {
		errMsg := fmt.Sprintf("连接Agent失败: %v", err)
		if resp != nil {
			errMsg = fmt.Sprintf("连接Agent失败: HTTP %d - %v", resp.StatusCode, err)
		}
		clientConn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		return
	}
	defer agentConn.Close()

	// 双向转发消息
	// 1. Agent -> Client
	go func() {
		for {
			messageType, message, err := agentConn.ReadMessage()
			if err != nil {
				return
			}
			if err := clientConn.WriteMessage(messageType, message); err != nil {
				return
			}
		}
	}()

	// 2. Client -> Agent
	for {
		messageType, message, err := clientConn.ReadMessage()
		if err != nil {
			return
		}
		if err := agentConn.WriteMessage(messageType, message); err != nil {
			return
		}
	}
}
