package feishu

import (
	"context"
	"fmt"

	"cmdb/config"
	"cmdb/pkg/middleware"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher/callback"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// CardActionHandler 卡片回调处理函数类型
type CardActionHandler func(ctx context.Context, event *callback.CardActionTriggerEvent) (*callback.CardActionTriggerResponse, error)

var cardActionHandlers []CardActionHandler

// RegisterCardActionHandler 注册卡片回调处理器
func RegisterCardActionHandler(handler CardActionHandler) {
	cardActionHandlers = append(cardActionHandlers, handler)
}

// StartWebSocketClient 启动飞书长连接客户端
func StartWebSocketClient() error {
	cfg := config.GetConfig().FeiShu
	if cfg.AppID == "" || cfg.AppSecret == "" {
		middleware.Logger.Warn("飞书配置不完整，跳过长连接初始化")
		return nil
	}

	// 注册事件处理器
	eventHandler := dispatcher.NewEventDispatcher(
		cfg.VerificationToken,
		cfg.EncryptKey,
	).OnP2CardActionTrigger(handleCardAction)

	// 创建长连接客户端
	cli := larkws.NewClient(cfg.AppID, cfg.AppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelInfo),
	)

	// 异步启动长连接
	go func() {
		middleware.Logger.Info("飞书长连接客户端启动中...")
		if err := cli.Start(context.Background()); err != nil {
			middleware.Logger.Errorf("飞书长连接启动失败: %v", err)
		}
	}()

	return nil
}

// handleCardAction 处理卡片回调
func handleCardAction(ctx context.Context, event *callback.CardActionTriggerEvent) (*callback.CardActionTriggerResponse, error) {
	middleware.Logger.Infof("收到飞书卡片回调: %s", larkcore.Prettify(event))

	// 遍历所有注册的处理器
	for _, handler := range cardActionHandlers {
		resp, err := handler(ctx, event)
		if err != nil {
			middleware.Logger.Errorf("卡片回调处理失败: %v", err)
			continue
		}
		if resp != nil {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("未找到匹配的卡片回调处理器")
}
