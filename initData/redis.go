package initData

import (
	"cmdb/types"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var redisClient *redis.Client

// RefreshRolePermissions 刷新角色的权限到Redis
func RefreshRolePermissions(roleID uint) error {
	permissions, err := getRolePermissions(roleID)
	if err != nil {
		return err
	}

	// 将权限列表存入Redis
	key := fmt.Sprintf("role:%d:permissions", roleID)
	return redisClient.Set(context.Background(), key, strings.Join(permissions, ","), types.TokenExpiration).Err()
}

// getRolePermissions 获取角色的所有权限标识
func getRolePermissions(roleID uint) ([]string, error) {
	db := GetDB()

	// 如果是超级管理员角色，返回所有权限
	if roleID == 1 { // 超级管理员
		var allPermissions []string
		err := db.Table("menus").
			Where("permission IS NOT NULL").
			Pluck("permission", &allPermissions).Error
		if err != nil {
			return nil, err
		}
		return allPermissions, nil
	}

	// 查询角色对应的权限
	var permissions []string
	err := db.Table("role_menus").
		Joins("JOIN menus ON menus.id = role_menus.menu_id").
		Where("role_menus.role_id = ? AND menus.permission IS NOT NULL AND menus.is_enabled = 1", roleID).
		Pluck("menus.permission", &permissions).Error
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// InitRolePermissions 初始化所有角色的权限到Redis
func InitRolePermissions() error {
	db := GetDB()

	// 获取所有角色ID
	var roleIDs []uint
	err := db.Table("roles").Pluck("id", &roleIDs).Error
	if err != nil {
		return fmt.Errorf("获取角色列表失败: %v", err)
	}

	// 初始化每个角色的权限到Redis
	for _, roleID := range roleIDs {
		err = RefreshRolePermissions(roleID)
		if err != nil {
			return fmt.Errorf("初始化角色[%d]权限失败: %v", roleID, err)
		}
	}

	return nil
}

// InitRedis 初始化Redis连接
func InitRedis(host string, port int, password string, db int) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("连接Redis失败: %v", err)
	}

	// 初始化角色权限到Redis
	err = InitRolePermissions()
	if err != nil {
		return fmt.Errorf("初始化角色权限失败: %v", err)
	}

	return nil
}

// GetRedis 获取Redis客户端实例
func GetRedis() *redis.Client {
	return redisClient
}

// SetToken 将用户token存入Redis
func SetToken(userID uint, token string, expiration time.Duration) error {
	key := fmt.Sprintf("user:%d:token", userID)
	return redisClient.Set(context.Background(), key, token, expiration).Err()
}

// GetToken 从Redis获取用户token
func GetToken(userID uint) (string, error) {
	key := fmt.Sprintf("user:%d:token", userID)
	return redisClient.Get(context.Background(), key).Result()
}

// DeleteToken 从Redis删除token
func DeleteToken(userID uint) error {
	key := fmt.Sprintf("token:%d", userID)
	return redisClient.Del(context.Background(), key).Err()
}
