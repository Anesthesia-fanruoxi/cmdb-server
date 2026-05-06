# 视图管理接口文档

## 1. 创建视图

### 接口描述
创建一个新的视图配置

### 请求信息
- 请求路径：`/api/elfk/view/create`
- 请求方法：POST
- 请求参数：
```json
{
    "name": "业务日志视图",        // 视图名称，必填
    "project": "cmdb",           // 项目标识，必填
    "index_pattern": "b_info-*", // 索引模式，必填
    "time_field": "@timestamp",  // 时间字段，必填
    "description": "业务相关日志"  // 视图描述，可选
}
```

### 响应信息
```json
{
    "code": 200,
    "message": "创建视图成功",
    "data": {
        "id": 1,
        "name": "业务日志视图",
        "project": "cmdb",
        "index_pattern": "b_info-*",
        "time_field": "@timestamp",
        "description": "业务相关日志",
        "create_time": "2024-03-19T10:00:00Z",
        "update_time": "2024-03-19T10:00:00Z"
    }
}
```

## 2. 更新视图

### 接口描述
更新现有视图的配置信息

### 请求信息
- 请求路径：`/api/elfk/view/update`
- 请求方法：PUT
- 请求参数：
```json
{
    "id": 1,                     // 视图ID，必填
    "name": "新的视图名称",       // 视图名称，可选
    "index_pattern": "new-*",    // 索引模式，可选
    "time_field": "log_time",    // 时间字段，可选
    "description": "新的描述"     // 视图描述，可选
}
```

### 响应信息
```json
{
    "code": 200,
    "message": "更新视图成功",
    "data": {
        "id": 1,
        "name": "新的视图名称",
        "project": "cmdb",
        "index_pattern": "new-*",
        "time_field": "log_time",
        "description": "新的描述",
        "create_time": "2024-03-19T10:00:00Z",
        "update_time": "2024-03-19T11:00:00Z"
    }
}
```

## 3. 删除视图

### 接口描述
删除指定的视图配置

### 请求信息
- 请求路径：`/api/elfk/view/delete`
- 请求方法：DELETE
- 请求参数：id（通过URL参数传递）
  例如：`/api/elfk/view/delete?id=1`

### 响应信息
```json
{
    "code": 200,
    "message": "删除视图成功",
    "data": null
}
```

## 4. 获取视图详情

### 接口描述
获取指定视图的详细信息

### 请求信息
- 请求路径：`/api/elfk/view/detail`
- 请求方法：GET
- 请求参数：id（通过URL参数传递）
  例如：`/api/elfk/view/detail?id=1`

### 响应信息
```json
{
    "code": 200,
    "message": "获取视图详情成功",
    "data": {
        "id": 1,
        "name": "业务日志视图",
        "project": "cmdb",
        "index_pattern": "b_info-*",
        "time_field": "@timestamp",
        "description": "业务相关日志",
        "create_time": "2024-03-19T10:00:00Z",
        "update_time": "2024-03-19T10:00:00Z"
    }
}
```

## 5. 获取视图列表

### 接口描述
获取当前用户有权限访问的所有视图列表

### 请求信息
- 请求路径：`/api/elfk/view/list`
- 请求方法：GET
- 请求参数：无

### 响应信息
```json
{
    "code": 200,
    "message": "获取视图列表成功",
    "data": [
        {
            "id": 1,
            "name": "业务日志视图",
            "project": "cmdb",
            "index_pattern": "b_info-*",
            "time_field": "@timestamp",
            "description": "业务相关日志",
            "create_time": "2024-03-19T10:00:00Z",
            "update_time": "2024-03-19T10:00:00Z"
        },
        // ... 更多视图
    ]
}
```

## 6. 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 无权限 |
| 404 | 视图不存在 |
| 500 | 服务器内部错误 |

## 7. 注意事项

1. 所有接口都需要进行用户认证和权限验证
2. 用户只能操作自己所属部门有权限的项目的视图
3. 同一项目下视图名称不能重复
4. 更新视图时，不允许修改所属项目
5. 时间字段必须是索引中存在的 date 类型字段
