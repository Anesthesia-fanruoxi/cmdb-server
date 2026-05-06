# 字典管理接口文档

## 1. 获取字典列表

获取所有字典记录列表。

### 请求信息
- 请求路径：`/api/system/dict/list`
- 请求方法：GET
- 权限要求：`system:dict:r`

### 响应信息
```json
{
    "code": 200,
    "message": "获取字典列表成功",
    "data": [
        {
            "id": 1,
            "dictName": "项目字典",
            "targetTable": "sys_project_dict",
            "keyName": "project",
            "valueName": "project_name",
            "createdBy": 1,
            "createdAt": "2024-02-19T10:00:00Z"
        }
    ]
}
```

## 2. 获取字典详情

获取指定字典的详细信息，包括字典记录和所有字典项。

### 请求信息
- 请求路径：`/api/system/dict/detail`
- 请求方法：GET
- 权限要求：`system:dict:r`

### 请求参数
| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| target_table | string | 是 | 目标表名 | sys_project_dict |

### 响应信息
```json
{
    "code": 200,
    "message": "获取字典详情成功",
    "data": {
        "dict": {
            "id": 1,
            "dictName": "项目字典",
            "targetTable": "sys_project_dict",
            "keyName": "project",
            "valueName": "project_name",
            "createdBy": 1,
            "createdAt": "2024-02-19T10:00:00Z"
        },
        "items": [
            {
                "key": "bd",
                "value": "百度"
            }
        ]
    }
}
```

## 3. 创建字典项

在指定的字典表中添加新的键值对。

### 请求信息
- 请求路径：`/api/system/dict/create`
- 请求方法：POST
- 权限要求：`system:dict:w`

### 请求参数
```json
{
    "tableName": "sys_project_dict",
    "key": "bd",
    "value": "百度",
    "color": "#409EFF"
}
```

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| tableName | string | 是 | 目标表名 | sys_project_dict |
| key | string | 是 | 键 | bd |
| value | string | 是 | 值 | 百度 |
| color | string | 否 | 显示颜色 | #409EFF |

### 响应信息
```json
{
    "code": 200,
    "message": "创建字典项成功",
    "data": {
        "project": "bd",
        "project_name": "百度"
    }
}
```

## 4. 删除字典项

删除指定字典表中的键值对。

### 请求信息
- 请求路径：`/api/system/dict/delete`
- 请求方法：DELETE
- 权限要求：`system:dict:w`

### 请求参数
| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| target_table | string | 是 | 目标表名 | sys_project_dict |
| key | string | 是 | 要删除的键 | bd |

### 响应信息
```json
{
    "code": 200,
    "message": "删除字典项成功"
}
```

## 5. 更新字典项

更新指定字典表中的键值对信息。

### 请求信息
- 请求路径：`/api/system/dict/update`
- 请求方法：PUT
- 权限要求：`system:dict:w`

### 请求参数
```json
{
    "target_table": "sys_project_dict",
    "key": "bd",
    "value": "百度集团",
    "color": "#FF5733"
}
```

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| target_table | string | 是 | 目标表名 | sys_project_dict |
| key | string | 是 | 键名 | bd |
| value | string | 否 | 新的值 | 百度集团 |
| color | string | 否 | 显示颜色 | #FF5733 |

### 响应信息
```json
{
    "code": 200,
    "message": "更新字典项成功"
}
```

注意：
1. value 和 color 字段可选，只更新提供的字段
2. 如果要修改 key，需要在 value 中提供新的键名
3. 颜色值支持标准的十六进制颜色代码

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 注意事项

1. 所有接口都需要登录认证
2. 创建字典项时，键不能重复
3. 删除字典项操作不可恢复
4. 字典表名必须存在于字典记录表中
