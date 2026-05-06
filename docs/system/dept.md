# 部门管理接口文档

## 1. 获取部门列表

获取所有部门的树形结构列表。

### 请求信息
- 请求路径：`/api/system/dept/list`
- 请求方法：GET
- 权限要求：`system:dept:r`

### 响应信息
```json
{
    "code": 200,
    "message": "查询成功",
    "data": [
        {
            "id": 1,
            "name": "总公司",
            "code": "HQ",
            "parent_id": 0,
            "description": "总部",
            "sort": 1,
            "children": [
                {
                    "id": 2,
                    "name": "研发部",
                    "code": "DEV",
                    "parent_id": 1,
                    "description": "研发部门",
                    "sort": 1,
                    "children": []
                }
            ]
        }
    ]
}
```

## 2. 获取部门详情

获取指定部门的详细信息。

### 请求信息
- 请求路径：`/api/system/dept/detail`
- 请求方法：GET
- 权限要求：`system:dept:r`

### 请求参数
| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| id | string | 是 | 部门ID | 1 |

### 响应信息
```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "id": 1,
        "name": "总公司",
        "code": "HQ",
        "parent_id": 0,
        "description": "总部",
        "sort": 1
    }
}
```

## 3. 创建部门

创建新的部门。

### 请求信息
- 请求路径：`/api/system/dept/create`
- 请求方法：POST
- 权限要求：`system:dept:w`

### 请求参数
```json
{
    "name": "研发部",
    "code": "DEV",
    "parent_id": 1,
    "description": "研发部门",
    "sort": 1
}
```

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| name | string | 是 | 部门名称 | 研发部 |
| code | string | 是 | 部门编码 | DEV |
| parent_id | int64 | 否 | 父部门ID | 1 |
| description | string | 否 | 部门描述 | 研发部门 |
| sort | int | 否 | 排序值 | 1 |

### 响应信息
```json
{
    "code": 200,
    "message": "创建成功",
    "data": null
}
```

## 4. 更新部门

更新部门信息。

### 请求信息
- 请求路径：`/api/system/dept/update`
- 请求方法：PUT
- 权限要求：`system:dept:w`

### 请求参数
```json
{
    "id": 1,
    "name": "研发中心",
    "code": "RD",
    "parent_id": 1,
    "description": "研发中心",
    "sort": 1
}
```

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| id | int64 | 是 | 部门ID | 1 |
| name | string | 否 | 部门名称 | 研发中心 |
| code | string | 否 | 部门编码 | RD |
| parent_id | int64 | 否 | 父部门ID | 1 |
| description | string | 否 | 部门描述 | 研发中心 |
| sort | int | 否 | 排序值 | 1 |

### 响应信息
```json
{
    "code": 200,
    "message": "更新成功",
    "data": null
}
```

## 5. 删除部门

删除指定的部门，同时处理相关的用户和子部门。

### 请求信息
- 请求路径：`/api/system/dept/delete`
- 请求方法：DELETE
- 权限要求：`system:dept:w`

### 请求参数
| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| id | string | 是 | 部门ID | 1 |

### 业务逻辑
1. 获取要删除部门及其所有子部门的ID列表
2. 将这些部门下的所有用户的部门ID修改为99（默认部门）
3. 删除这些部门在 sys_dept_projects 表中的关联关系
4. 删除这些部门的记录

### 响应信息
```json
{
    "code": 200,
    "message": "删除成功",
    "data": null
}
```

### 特殊说明
- 删除部门时会自动处理该部门及其子部门下的所有用户
- 用户不会被删除，而是被转移到ID为99的默认部门
- 如果部门下有子部门，会递归处理所有子部门
- 删除部门时会同时清除该部门的所有项目关联关系

## 6. 获取部门项目关联

获取部门关联的项目列表。

### 请求信息
- 请求路径：`/api/system/dept/project/detail`
- 请求方法：GET
- 权限要求：`system:dept:r`

### 请求参数
| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| dept_id | string | 是 | 部门ID | 1 |

### 响应信息
```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "dept_id": 1,
        "project": ["project1", "project2"]
    }
}
```

## 7. 更新部门项目关联

更新部门关联的项目列表。

### 请求信息
- 请求路径：`/api/system/dept/project/update`
- 请求方法：POST
- 权限要求：`system:dept:w`

### 请求参数
```json
{
    "dept_id": 1,
    "project": ["project1", "project2"]
}
```

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| dept_id | int64 | 是 | 部门ID | 1 |
| project | []string | 是 | 项目代码列表 | ["project1", "project2"] |

### 响应信息
```json
{
    "code": 200,
    "message": "更新成功",
    "data": null
}
```

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
2. 部门名称和编码在同级部门中必须唯一
3. 删除部门时会同时删除其下所有子部门，相关用户会被转移到默认部门（ID=99）
4. 更新部门时不能将父部门设置为自己或自己的子部门
5. 部门项目关联更新时会覆盖原有关联关系
