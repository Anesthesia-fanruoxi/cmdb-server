package path

const (
	// Redis查询
	RedisProjectsPath = "/api/redis/projects" // 获取可访问的项目列表
	RedisInfoPath     = "/api/redis/info"     // 获取Redis信息
	RedisTreePath     = "/api/redis/tree"     // 获取Redis keys树形列表
	RedisKeyPath      = "/api/redis/key"      // 获取Redis key的值
	RedisDeletePath   = "/api/redis/delete"   // 删除Redis key
)
