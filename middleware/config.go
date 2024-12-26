package middleware

// 白名单路径
var WhiteList = map[string]bool{
	"/api/user/login":  true, // 用户登录
	"/api/user/create": true, // 用户注册
}

// CheckWhiteList 检查路径是否在白名单中
func CheckWhiteList(path string) bool {
	return WhiteList[path]
}
