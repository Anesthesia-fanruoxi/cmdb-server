# 用户管理接口文档

## 个人信息管理

### 1. 获取个人信息
#### 接口描述
获取当前登录用户的个人信息

#### 请求信息
- 请求路径：`/api/system/user/profile`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

#### 响应示例
```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "id": 1,
        "user_name": "admin",
        "nick_name": "管理员",
        "email": "admin@example.com",
        "phone": "13800138000",
        "role_id": 1,
        "role_name": "超级管理员",
        "dept_id": 1,
        "dept_name": "运维部门",
        "is_enabled": true,
        "created_at": "2024-12-23T14:29:24.193+08:00",
        "updated_at": "2025-02-17T11:58:08.771+08:00"
    }
}
```

### 2. 更新个人信息
#### 接口描述
更新当前登录用户的个人信息

#### 请求信息
- 请求路径：`/api/system/user/profile`
- 请求方法：PUT
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

#### 请求参数
| 参数名   | 类型   | 必填 | 描述     | 验证规则 |
|----------|--------|------|----------|-----------|
| nick_name | string | 否 | 昵称     | 最大32字符|
| email    | string | 否 | 邮箱     | 有效邮箱  |
| phone    | string | 否  | 手机号   | 11位      |
| password | string | 否   | 新密码   | 密码规则  |

#### 响应示例
```json
{
    "code": 200,
    "message": "更新成功"
}
```

## 用户管理接口

## 1. 用户登录
### 接口描述
用户登录系统，获取访问令牌。支持密码登录和TOTP验证码登录两种方式。

### 请求信息
- 请求路径：`/api/system/user/login`
- 请求方法：POST
- 请求头：Content-Type: application/json

### 请求参数
| 参数名    | 类型   | 必填 | 描述                              |
|-----------|--------|------|---------------------------------|
| user_name | string | 是   | 用户名                           |
| login_type| string | 是   | 登录类型，可选值：`password`(密码登录)、`totp`(验证码登录) |
| password  | string | 否   | 密码（login_type为password时必填） |
| totp_code | string | 否   | TOTP验证码（login_type为totp时必填） |

### 响应参数
| 参数名    | 类型   | 描述         |
|-----------|--------|--------------|
| token     | string | 访问令牌     |
| user_id   | int64  | 用户ID       |
| user_name  | string | 用户名       |
| role_id   | int64  | 角色ID       |
| role_name | string | 角色名称     |

### 响应示例
```json
{
    "code": 200,
    "message": "登录成功",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "user_id": 1,
        "user_name": "admin",
        "role_id": 1,
        "role_name": "超级管理员"
    }
}
```

### 示例请求 - 密码登录
```json
{
    "user_name": "admin",
    "login_type": "password",
    "password": "your_password"
}
```

### 示例请求 - TOTP验证码登录
```json
{
    "user_name": "admin",
    "login_type": "totp",
    "totp_code": "123456"
}
```

## 2. 创建用户
### 接口描述
创建新用户

### 请求信息
- 请求路径：`/api/system/user/create`
- 请求方法：POST
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名   | 类型   | 必填 | 描述     | 验证规则 |
|----------|--------|------|----------|-----------|
| user_name | string | 是   | 用户名   | 3-32字符  |
| password | string | 是   | 密码     | 密码规则  |
| nick_name | string | 是   | 昵称     | 最大32字符|
| email    | string | 是   | 邮箱     | 有效邮箱  |
| phone    | string | 是   | 手机号   | 11位      |
| role_id  | int64  | 是   | 角色ID   | >0        |
| dept_id  | int64  | 是   | 部门ID   | >0        |

### 响应示例
```json
{
    "code": 200,
    "message": "用户创建成功",
    "data": {
        "user_id": 1
    }
}
```

## 3. 更新用户
### 接口描述
更新用户信息

### 请求信息
- 请求路径：`/api/system/user/update`
- 请求方法：PUT
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名     | 类型    | 必填 | 描述       |
|------------|---------|------|------------|
| id         | int64   | 否   | 用户ID     |
| password   | string  | 否   | 新密码     |
| nick_name   | string  | 否   | 昵称       |
| email      | string  | 否   | 邮箱       |
| phone      | string  | 否   | 手机号     |
| is_enabled | boolean | 否   | 是否启用   |
| role_id    | int64   | 否   | 角色ID     |
| dept_id    | int64   | 否   | 部门ID     |

### 响应示例
```json
{
    "code": 200,
    "message": "更新成功"
}
```

## 4. 获取用户列表
### 接口描述
获取用户列表，支持分页和筛选

### 请求信息
- 请求路径：`/api/system/user/list`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 请求参数
| 参数名    | 类型   | 必填 | 描述         |
|-----------|--------|------|--------------|
| page      | int    | 否   | 页码，默认1  |
| page_size | int    | 否   | 每页数量，默认20 |
| user_name  | string | 否   | 用户名关键字 |
| nick_name  | string | 否   | 昵称关键字   |
| role_id   | int64  | 否   | 角色ID       |
| dept_id   | int64  | 否   | 部门ID       |

### 响应参数
| 参数名     | 类型    | 描述         |
|------------|---------|--------------|
| total      | int64   | 总记录数     |
| list       | array   | 用户列表     |
| page       | int     | 当前页码     |
| page_size  | int     | 每页数量     |

### 响应示例
```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "list": [
            {
                "id": 1,
                "user_name": "admin",
                "nick_name": "管理员",
                "email": "admin@example.com",
                "phone": "13800138000",
                "role_id": 1,
                "role_name": "超级管理员",
                "dept_id": 1,
                "dept_name": "运维部门",
                "is_enabled": true,
                "created_at": "2024-12-23T14:29:24.193+08:00",
                "updated_at": "2025-02-17T11:58:08.771+08:00"
            },
            {
                "id": 2,
                "user_name": "admin11",
                "nick_name": "管理员",
                "email": "admin@example.com",
                "phone": "13800138000",
                "role_id": 1,
                "role_name": "超级管理员",
                "dept_id": 1,
                "dept_name": "运维部门",
                "is_enabled": true,
                "created_at": "2025-02-14T17:05:27.948+08:00",
                "updated_at": "2025-02-14T17:05:27.948+08:00"
            }
        ],
        "page": 1,
        "page_size": 20,
        "total": 2
    }
}
```

## 5. 获取用户详情
### 接口描述
获取指定用户的详细信息

### 请求信息
- 请求路径：`/api/system/user/detail?{id}`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 响应示例
```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "id": 1,
        "role_id": 1,
        "is_enabled": true,
        "user_name": "admin",
        "created_at": "2024-12-23T14:29:24.193+08:00",
        "updated_at": "2025-02-17T11:58:08.771+08:00",
        "nick_name": "管理员",
        "email": "admin@example.com",
        "phone": "13800138000",
        "dept_id": 1,
        "role": {
            "id": 0,
            "created_at": "0001-01-01T00:00:00Z",
            "updated_at": "0001-01-01T00:00:00Z",
            "name": "",
            "code": "",
            "description": "",
            "level": 0
        }
    }
}
```

## 6. 删除用户
### 接口描述
删除指定用户

### 请求信息
- 请求路径：`/api/system/user/delete?{id}`
- 请求方法：DELETE
- 请求头：Authorization: Bearer {token}

### 响应示例
```json
{
    "code": 200,
    "message": "删除成功"
}
```

## 7. 用户登出
### 接口描述
修改用户密码

### 请求信息
- 请求路径：`/api/system/user/logout`
- 请求方法：POST
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

|      |        |      |        |
| ---- | ------ | ---- | ------ |
|      |        |      |        |
|      | string | 是   | 新密码 |

### 响应示例
```json
{
    "code": 200,
    "message": "登出成功"
}
```

## 8. 批量删除用户
### 接口描述
批量删除指定用户列表

### 请求信息
- 请求路径：`/api/system/user/batch-delete`
- 请求方法：POST
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名   | 类型     | 必填 | 描述         | 验证规则 |
|----------|----------|------|--------------|----------|
| user_ids | []uint   | 是   | 用户ID列表   | 非空数组 |

### 响应示例
```json
{
    "code": 200,
    "message": "删除成功"
}
```

### 错误响应
```json
{
    "code": 400,
    "message": "不能删除超级管理员用户"
}
```
```json
{
    "code": 403,
    "message": "无权删除其他部门的用户"
}
```

## 9. 批量更新用户
### 接口描述
批量更新指定用户列表的部门和状态信息

### 请求信息
- 请求路径：`/api/system/user/batch-update`
- 请求方法：PUT
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名        | 类型     | 必填 | 描述           | 说明 |
|---------------|----------|------|----------------|------|
| user_ids      | []uint   | 是   | 用户ID列表     | 非空数组 |
| dept_id       | uint     | 否   | 部门ID         | 更新用户部门 |
| is_enabled    | boolean  | 否   | 是否启用       | 更新用户状态 |
| online_assets | boolean  | 否   | 线上资产权限   | 更新线上资产权限 |
| test_assets   | boolean  | 否   | 测试资产权限   | 更新测试资产权限 |

### 响应示例
```json
{
    "code": 200,
    "message": "更新成功"
}
```

### 错误响应
```json
{
    "code": 400,
    "message": "至少需要更新一个字段"
}
```
```json
{
    "code": 400,
    "message": "不能修改超级管理员用户"
}
```
```json
{
    "code": 403,
    "message": "无权修改其他部门的用户"
}
```

## 权限说明
1. 超级管理员可以修改所有用户信息
2. 普通管理员只能修改同部门且角色级别低于自己的用户
3. 用户可以修改自己的基本信息（不包括角色和部门）
4. 不能删除或修改超级管理员用户
5. 不能删除自己的账号

## 错误码说明
| 错误码 | 描述              |
| ------ | ----------------- |
| 400    | 请求参数错误      |
| 401    | 未登录或token失效 |
| 403    | 无权限操作        |
| 404    | 用户不存在        |
| 500    | 服务器内部错误    |

## 注意事项
1. 密码必须符合复杂度要求
2. 邮箱和手机号需要符合格式要求
3. token 需要在请求头中携带
4. 用户名不能重复