# 角色管理接口文档

## 1. 获取角色列表
### 接口描述
获取系统中所有角色列表，不需要任何查询参数

### 请求信息
- 请求路径：`/api/system/role/list`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 响应参数
| 参数名      | 类型   | 描述     |
|-------------|--------|----------|
| id          | int64  | 角色ID   |
| name        | string | 角色名称 |
| code        | string | 角色编码 |
| description | string | 角色描述 |
| level       | int    | 角色级别 |
| created_at  | string | 创建时间 |
| updated_at  | string | 更新时间 |

### 响应示例
```json
{
    "code": 200,
    "message": "查询成功",
    "data": [
        {
            "id": 1,
            "name": "超级管理员",
            "code": "admin",
            "description": "系统超级管理员",
            "level": 0,
            "created_at": "2024-12-23T14:29:24.193+08:00",
            "updated_at": "2024-12-23T14:29:24.193+08:00"
        }
    ]
}
```

## 2. 创建角色
### 接口描述
创建新的角色

### 请求信息
- 请求路径：`/api/system/role/create`
- 请求方法：POST
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名      | 类型   | 必填 | 描述     | 验证规则 |
|-------------|--------|------|----------|-----------|
| name        | string | 是   | 角色名称 | 最大32字符|
| code        | string | 是   | 角色编码 | 最大32字符|
| description | string | 否   | 角色描述 | 最大128字符|
| level       | int    | 是   | 角色级别 | >=0       |

### 响应示例
```json
{
    "code": 200,
    "message": "创建成功",
    "data": {
        "role_id": 1
    }
}
```

## 3. 更新角色
### 接口描述
更新角色信息

### 请求信息
- 请求路径：`/api/system/role/update`
- 请求方法：PUT
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名      | 类型   | 必填 | 描述     |
|-------------|--------|------|----------|
| id          | int64  | 是   | 角色ID   |
| name        | string | 否   | 角色名称 |
| code        | string | 否   | 角色编码 |
| description | string | 否   | 角色描述 |
| level       | int    | 否   | 角色级别 |

### 响应示例
```json
{
    "code": 200,
    "message": "更新成功"
}
```

## 4. 删除角色
### 接口描述
删除指定角色

### 请求信息
- 请求路径：`/api/system/role/delete?id={id}`
- 请求方法：DELETE
- 请求头：Authorization: Bearer {token}

### 响应示例
```json
{
    "code": 200,
    "message": "删除成功"
}
```

## 5. 获取角色菜单权限
### 接口描述
获取指定角色的菜单权限

### 请求信息
- 请求路径：`/api/system/role/menus/detail?role_id={role_id}`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 响应示例
```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "9": [
            1,
            0,
            0
        ],
        "91": [
            1,
            1,
            1
        ],
        "92": [
            1,
            1,
            1
        ],
        "93": [
            1,
            1,
            1
        ],
        "94": [
            1,
            1,
            1
        ],
        "95": [
            1,
            1,
            1
        ]
    }
}
```

## 6. 更新角色菜单权限
### 接口描述
更新指定角色的菜单权限

### 请求信息
- 请求路径：`/api/system/role/menu`
- 请求方法：PUT
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名      | 类型    | 必填 | 描述       |
| ----------- | ------- | ---- | ---------- |
| role_id     | int64   | 是   | 角色ID     |
| permissions | []int64 | 是   | 菜单ID列表 |

### 请求示例

```json
{
    "role_id": 1,
    "permissions": 
        {
            "9": [1,0,0],
            "91": [1,1,1],
            "92": [1,1,1],
            "93": [1,1,1],
            "94": [1,1,1],
            "95": [1,1,1]
        }
}
```



### 响应示例

```json
{
    "code": 200,
    "message": "更新成功"
}
```

## 权限说明
1. 只有超级管理员可以管理角色
2. 角色级别为0的是超级管理员角色，不能修改或删除
3. 不能删除已分配给用户的角色

## 错误码说明
| 错误码 | 描述                           |
|--------|--------------------------------|
| 400    | 请求参数错误                   |
| 401    | 未登录或token失效              |
| 403    | 无权限操作                     |
| 404    | 角色不存在                     |
| 409    | 角色编码或名称重复             |
| 500    | 服务器内部错误                 |

## 注意事项
1. 角色编码必须唯一
2. 角色级别越小权限越大，0为最高级别
3. 删除角色前需要确保没有用户使用该角色
4. 更新角色菜单权限会立即生效
