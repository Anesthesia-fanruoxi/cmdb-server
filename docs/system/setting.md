# 系统设置接口文档

## 获取系统设置

获取系统的基本配置信息。

### 请求信息

- 请求路径: `/api/system/setting/basic`
- 请求方法: `GET`
- 需要权限: `system:setting:r`

### 响应信息

#### 成功响应

```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": 1,
        "system_name": "CMDB资产管理系统",
        "system_short_name": "CMDB",
        "password_min_length": 8,
        "password_max_length": 32,
        "password_need_number": true,
        "password_need_letter": true,
        "password_need_case": true,
        "password_need_special": false,
        "login_protection": true,
        "login_fail_count": 5,
        "login_lock_time": 30,
        "default_password": "Cmdb@123",
        "token_expire_time": 24,
        "created_at": "2024-02-18T10:00:00+08:00",
        "updated_at": "2024-02-18T10:00:00+08:00"
    }
}
```

#### 错误响应

```json
{
    "code": 500,
    "message": "获取系统设置失败",
    "data": null
}
```

## 更新系统设置

更新系统的基本配置信息。

### 请求信息

- 请求路径: `/api/system/setting/basic/update`
- 请求方法: `POST`
- 需要权限: `system:setting:w`

### 请求参数

```json
{
    "id": 1,
    "system_name": "CMDB资产管理系统",
    "system_short_name": "CMDB",
    "password_min_length": 8,
    "password_max_length": 32,
    "password_need_number": true,
    "password_need_letter": true,
    "password_need_case": true,
    "password_need_special": false,
    "login_protection": true,
    "login_fail_count": 5,
    "login_lock_time": 30,
    "default_password": "Cmdb@123",
    "token_expire_time": 24
}
```

### 参数说明

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int | 是 | 设置ID |
| system_name | string | 是 | 系统名称 |
| system_short_name | string | 是 | 系统简称 |
| password_min_length | int | 否 | 密码最小长度，默认8 |
| password_max_length | int | 否 | 密码最大长度，默认32 |
| password_need_number | bool | 否 | 密码是否需要数字，默认true |
| password_need_letter | bool | 否 | 密码是否需要字母，默认true |
| password_need_case | bool | 否 | 密码是否需要大小写，默认true |
| password_need_special | bool | 否 | 密码是否需要特殊字符，默认false |
| login_protection | bool | 否 | 是否开启登录保护，默认true |
| login_fail_count | int | 否 | 登录失败次数，默认5次 |
| login_lock_time | int | 否 | 登录锁定时间(分钟)，默认30分钟 |
| default_password | string | 否 | 默认密码，默认"Cmdb@123" |
| token_expire_time | int | 否 | token过期时间(小时)，默认24小时 |

### 响应信息

#### 成功响应

```json
{
    "code": 200,
    "message": "更新成功",
    "data": {
        "id": 1,
        "system_name": "CMDB资产管理系统",
        "system_short_name": "CMDB",
        "password_min_length": 8,
        "password_max_length": 32,
        "password_need_number": true,
        "password_need_letter": true,
        "password_need_case": true,
        "password_need_special": false,
        "login_protection": true,
        "login_fail_count": 5,
        "login_lock_time": 30,
        "default_password": "Cmdb@123",
        "token_expire_time": 24,
        "created_at": "2024-02-18T10:00:00+08:00",
        "updated_at": "2024-02-18T10:00:00+08:00"
    }
}
```

#### 错误响应

```json
{
    "code": 400,
    "message": "无效的请求参数",
    "data": null
}
```

或

```json
{
    "code": 500,
    "message": "更新系统设置失败",
    "data": null
}
```