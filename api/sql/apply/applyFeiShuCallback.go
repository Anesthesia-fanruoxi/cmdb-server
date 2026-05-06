package apply

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	models "cmdb/models/sql"
	"cmdb/pkg/database"
	"cmdb/pkg/middleware"

	"gorm.io/gorm"
)

// FeiShuCardCallback 飞书卡片回调请求结构
type FeiShuCardCallback struct {
	OpenID        string           `json:"open_id"`         // 点击用户的open_id
	UserID        string           `json:"user_id"`         // 点击用户的user_id
	OpenMessageID string           `json:"open_message_id"` // 消息ID
	TenantKey     string           `json:"tenant_key"`      // 租户key
	Token         string           `json:"token"`           // 用于更新卡片的token
	Action        FeiShuCardAction `json:"action"`          // 按钮携带的数据
	Challenge     string           `json:"challenge"`       // challenge验证（首次配置时使用）
}

// FeiShuCardAction 卡片交互动作
type FeiShuCardAction struct {
	Value map[string]string `json:"value"` // 按钮携带的value
	Tag   string            `json:"tag"`   // 交互组件类型
}

// FeiShuApplyCallback 飞书卡片回调接口 - SQL执行审批
func FeiShuApplyCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		middleware.ResponseError(w, http.StatusMethodNotAllowed, "只支持POST请求")
		return
	}

	var req FeiShuCardCallback
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ResponseError(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 处理challenge验证（首次配置回调地址时）
	if req.Challenge != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"challenge": req.Challenge})
		return
	}

	middleware.Logger.Infof("飞书卡片回调 - UserID: %s, OpenID: %s, Action: %+v",
		req.UserID, req.OpenID, req.Action.Value)

	// 解析action中的业务数据
	applyIDStr := req.Action.Value["apply_id"]
	action := req.Action.Value["action"]

	if applyIDStr == "" || action == "" {
		respondWithCard(w, createErrorCard("参数错误", "缺少必要的业务参数"))
		return
	}

	applyID, err := strconv.ParseInt(applyIDStr, 10, 64)
	if err != nil {
		respondWithCard(w, createErrorCard("参数错误", "审批ID格式错误"))
		return
	}

	// 获取流程信息
	var apply models.SQLApply
	if err := database.DB.Where("id = ?", applyID).First(&apply).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithCard(w, createErrorCard("流程不存在", "该审批流程不存在或已被删除"))
		} else {
			respondWithCard(w, createErrorCard("系统错误", "查询流程失败"))
		}
		return
	}

	// 验证用户身份 - 通过user_id匹配系统用户
	operatorID, err := getUserIDByFeiShuID(req.UserID, req.OpenID)
	if err != nil {
		middleware.Logger.Warnf("飞书用户身份验证失败: %v", err)
		respondWithCard(w, createErrorCard("身份验证失败", "无法识别您的身份，请确认飞书账号已关联系统"))
		return
	}

	// 验证操作权限
	if err := validateCallbackPermission(&apply, operatorID, action); err != nil {
		respondWithCard(w, createErrorCard("权限不足", err.Error()))
		return
	}

	// 使用分布式锁防止重复操作
	lockKey := fmt.Sprintf("sql_apply_callback:%d", applyID)
	success, err := database.RDB.SetNX(context.Background(), lockKey, "1", 30*time.Second).Result()
	if err != nil || !success {
		respondWithCard(w, createErrorCard("操作冲突", "该流程正在处理中，请稍后再试"))
		return
	}
	defer database.RDB.Del(context.Background(), lockKey)

	// 执行审批操作
	resultCard, err := executeApplyAction(&apply, action)
	if err != nil {
		middleware.Logger.Errorf("执行审批操作失败: %v", err)
		respondWithCard(w, createErrorCard("操作失败", err.Error()))
		return
	}

	middleware.Logger.Infof("飞书卡片回调成功 - ApplyID: %d, Action: %s, Operator: %d",
		applyID, action, operatorID)

	respondWithCard(w, resultCard)
}
