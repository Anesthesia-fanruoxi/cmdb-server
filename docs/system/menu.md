# 菜单管理接口文档

## 1. 获取菜单列表
### 接口描述
获取系统中所有菜单列表，以树形结构返回

### 请求信息
- 请求路径：`/api/system/menu/list`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 响应参数
| 参数名      | 类型   | 描述     |
|-------------|--------|----------|
| id          | int64  | 菜单ID   |
| parent_id   | int64  | 父菜单ID |
| name        | string | 菜单名称 |
| path        | string | 路由路径 |
| component   | string | 组件路径 |
| sort        | int    | 排序     |
| icon        | string | 图标     |
| is_show     | bool   | 是否显示 |
| children    | array  | 子菜单   |

### 响应示例
```json
{
    "code": 200,
    "message": "查询成功",
    "data": [
        {
            "id": 1,
            "parent_id": 0,
            "name": "系统管理",
            "path": "/system",
            "component": "Layout",
            "sort": 1,
            "icon": "setting",
            "is_show": true,
            "children": [
                {
                    "id": 2,
                    "parent_id": 1,
                    "name": "用户管理",
                    "path": "user",
                    "component": "system/user/index",
                    "sort": 1,
                    "icon": "user",
                    "is_show": true
                }
            ]
        }
    ]
}
```

## 2. 创建菜单
### 接口描述
创建新的菜单

### 请求信息
- 请求路径：`/api/system/menu/create`
- 请求方法：POST
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名    | 类型   | 必填 | 描述     | 验证规则 |
|-----------|--------|------|----------|-----------|
| parent_id | int64  | 是   | 父菜单ID | >=0      |
| name      | string | 是   | 菜单名称 | 最大32字符|
| path      | string | 是   | 路由路径 | 最大128字符|
| component | string | 是   | 组件路径 | 最大128字符|
| sort      | int    | 否   | 排序     | >=0      |
| icon      | string | 否   | 图标     | 最大32字符|
| is_show   | bool   | 否   | 是否显示 | -        |

### 请求示例

```json
{
    "parent_id": 9,
    "name": "测试管理",
    "path": "/system/ceshi",
    "component": "system/dept/index",
    "permission": "system:dept",
    "sort": 5,
    "icon": "OfficeBuilding",
    "is_visible": true,
    "is_enabled": true
}
```



### 响应示例

```json
{
    "code": 200,
    "message": "创建成功",
    "data": {
        "menu_id": 1
    }
}
```

## 3. 更新菜单
### 接口描述
更新菜单信息

### 请求信息
- 请求路径：`/api/system/menu/update`
- 请求方法：PUT
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名    | 类型   | 必填 | 描述     |
|-----------|--------|------|----------|
| id        | int64  | 是   | 菜单ID   |
| parent_id | int64  | 否   | 父菜单ID |
| name      | string | 否   | 菜单名称 |
| path      | string | 否   | 路由路径 |
| component | string | 否   | 组件路径 |
| sort      | int    | 否   | 排序     |
| icon      | string | 否   | 图标     |
| is_show   | bool   | 否   | 是否显示 |

### 响应示例
```json
{
    "code": 200,
    "message": "更新成功"
}
```

## 4. 删除菜单
### 接口描述
删除指定菜单及其子菜单

### 请求信息
- 请求路径：`/api/system/menu/delete?id={id}`
- 请求方法：DELETE
- 请求头：Authorization: Bearer {token}

### 响应示例
```json
{
    "code": 200,
    "message": "删除成功"
}
```

## 5. 获取用户菜单列表
### 接口描述
获取当前登录用户的可访问菜单列表

### 请求信息
- 请求路径：`/api/system/menu/user`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 响应参数
| 参数名      | 类型   | 描述     |
|-------------|--------|----------|
| id          | int64  | 菜单ID   |
| parent_id   | int64  | 父菜单ID |
| name        | string | 菜单名称 |
| path        | string | 路由路径 |
| component   | string | 组件路径 |
| sort        | int    | 排序     |
| icon        | string | 图标     |
| is_enabled  | bool   | 是否启用 |
| children    | array  | 子菜单   |

### 响应示例
```json
{
    "code": 200,
    "message": "查询成功",
    "data": [
        {
            "id": 1,
            "parent_id": 0,
            "name": "系统管理",
            "path": "/system",
            "component": "Layout",
            "sort": 1,
            "icon": "setting",
            "is_enabled": true,
            "children": [
                {
                    "id": 2,
                    "parent_id": 1,
                    "name": "用户管理",
                    "path": "user",
                    "component": "system/user/index",
                    "sort": 1,
                    "icon": "user",
                    "is_enabled": true
                }
            ]
        }
    ]
}
```

## 6. 获取菜单树
### 接口描述
获取所有菜单的树形结构，用于菜单管理

### 请求信息
- 请求路径：`/api/system/menu/tree`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 响应参数
| 参数名      | 类型   | 描述     |
|-------------|--------|----------|
| id          | int64  | 菜单ID   |
| parent_id   | int64  | 父菜单ID |
| name        | string | 菜单名称 |
| path        | string | 路由路径 |
| component   | string | 组件路径 |
| sort        | int    | 排序     |
| icon        | string | 图标     |
| is_enabled  | bool   | 是否启用 |
| children    | array  | 子菜单   |

### 响应示例
```json
{
    "code": 200,
    "message": "查询成功",
    "data": [
        {
            "id": 1,
            "parent_id": 0,
            "name": "系统管理",
            "path": "/system",
            "component": "Layout",
            "sort": 1,
            "icon": "setting",
            "is_enabled": true,
            "children": [
                {
                    "id": 2,
                    "parent_id": 1,
                    "name": "用户管理",
                    "path": "user",
                    "component": "system/user/index",
                    "sort": 1,
                    "icon": "user",
                    "is_enabled": true,
                    "children": []
                },
                {
                    "id": 3,
                    "parent_id": 1,
                    "name": "角色管理",
                    "path": "role",
                    "component": "system/role/index",
                    "sort": 2,
                    "icon": "peoples",
                    "is_enabled": true,
                    "children": []
                }
            ]
        }
    ]
}
```

### 区别说明
1. `/api/system/menu/user` 接口返回当前用户有权限访问的菜单
2. `/api/system/menu/tree` 接口返回所有菜单，用于管理员管理菜单
3. `/api/system/menu/list` 接口返回扁平的菜单列表，不包含树形结构 

## 权限说明
1. 只有超级管理员可以管理菜单
2. 删除菜单时会同时删除其子菜单
3. 删除菜单时会同时删除相关的角色菜单权限关联

## 错误码说明
| 错误码 | 描述              |
| ------ | ----------------- |
| 400    | 请求参数错误      |
| 401    | 未登录或token失效 |
| 403    | 无权限操作        |
| 404    | 菜单不存在        |
| 409    | 菜单路径重复      |
| 500    | 服务器内部错误    |

## 注意事项
1. 菜单路径必须唯一
2. 创建子菜单时需确保父菜单存在
3. 更新菜单时不能将菜单设置为自己的子菜单
4. 菜单排序值越小越靠前 