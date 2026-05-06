# CMDB 审计日志接口文档

## 概述

CMDB 系统的审计日志模块提供了对系统操作和用户登录行为的记录和查询功能。通过审计日志，管理员可以追踪系统中的各种操作，包括用户登录、资源创建、修改和删除等活动，有助于系统安全管理和问题排查。

## 审计日志类型

系统支持三种类型的审计日志：

1. **操作日志**：记录用户对系统资源的操作，如创建、修改、删除等。
2. **登录日志**：记录用户的登录行为，包括登录成功和失败的情况。
3. **ES查询日志**：记录用户对Elasticsearch的查询操作，包括查询条件、响应时间和结果数量等。

## 数据库表结构

### 操作日志表 (audit_operation_log)

```sql
CREATE TABLE `audit_operation_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `user_name` varchar(50) NOT NULL COMMENT '用户名',
  `module` varchar(50) NOT NULL COMMENT '操作模块',
  `action` varchar(50) NOT NULL COMMENT '操作类型',
  `method` varchar(10) NOT NULL COMMENT '请求方法',
  `url` varchar(255) NOT NULL COMMENT '请求URL',
  `ip` varchar(50) NOT NULL COMMENT '操作IP',
  `request` text COMMENT '请求参数',
  `response` text COMMENT '响应结果',
  `status` int NOT NULL COMMENT '操作状态(0:失败,1:成功)',
  `error_msg` text COMMENT '错误消息',
  `duration` int NOT NULL COMMENT '执行时长(毫秒)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='操作日志表';
```

### 登录日志表 (audit_login_log)

```sql
CREATE TABLE `audit_login_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `user_name` varchar(50) NOT NULL COMMENT '用户名',
  `ip` varchar(50) NOT NULL COMMENT '登录IP',
  `location` varchar(255) DEFAULT NULL COMMENT '登录地点',
  `browser` varchar(50) DEFAULT NULL COMMENT '浏览器类型',
  `os` varchar(50) DEFAULT NULL COMMENT '操作系统',
  `status` tinyint NOT NULL COMMENT '登录状态(0:失败,1:成功)',
  `msg` varchar(255) DEFAULT NULL COMMENT '提示消息',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='登录日志表';
```

### ES查询日志表 (sys_es_log)

```sql
CREATE TABLE `sys_es_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` VARCHAR(64) COMMENT '用户ID',
  `username` VARCHAR(100) COMMENT '用户名',
  `client_ip` VARCHAR(64) COMMENT '客户端IP',
  `project` VARCHAR(255) COMMENT '项目名称',
  `view_name` VARCHAR(255) COMMENT '视图名称',
  `q_index_pattern` VARCHAR(255) COMMENT '查询索引模式',
  `q_time_field` VARCHAR(255) COMMENT '查询时间字段',
  `q_start_time` VARCHAR(255) COMMENT '查询开始时间',
  `q_end_time` VARCHAR(255) COMMENT '查询结束时间',
  `q_keyword` VARCHAR(255) COMMENT '搜索关键词',
  `response_code` INT COMMENT '响应状态码',
  `response_body` TEXT COMMENT '响应体内容',
  `query_time_ms` INT COMMENT '查询耗时(毫秒)',
  `doc_count` INT COMMENT '查询文档数量',
  `error_message` TEXT COMMENT '错误信息',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_project` (`project`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_view_name` (`view_name`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Elasticsearch审计日志表';
```

## API 接口

### 1. 查询操作日志

#### 请求信息

- **URL**: `/api/audit/operation`
- **方法**: `POST`
- **权限要求**: 需要管理员权限

#### 请求参数

| 参数名    | 类型     | 必填 | 描述                 |
|-----------|----------|------|----------------------|
| userName  | string   | 否   | 用户名（支持模糊查询）|
| module    | string   | 否   | 操作模块             |
| action    | string   | 否   | 操作类型             |
| startTime | datetime | 否   | 开始时间             |
| endTime   | datetime | 否   | 结束时间             |
| page      | int      | 否   | 页码（默认为1）      |
| pageSize  | int      | 否   | 每页记录数（默认20） |

#### 请求示例

```json
{
  "userName": "admin",
  "module": "user",
  "action": "create",
  "page": 1,
  "pageSize": 20
}
```

#### 响应参数

| 参数名     | 类型   | 描述               |
|------------|--------|-------------------|
| code       | int    | 状态码，200表示成功 |
| message    | string | 响应消息           |
| data       | object | 响应数据           |
| data.total | int    | 总记录数           |
| data.list  | array  | 操作日志列表       |
| data.page  | int    | 当前页码           |
| data.pageSize | int  | 每页记录数         |
| data.totalPages | int | 总页数           |

#### 响应示例

```json
{
  "code": 200,
  "message": "请求成功",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "userId": 1,
        "userName": "admin",
        "module": "user",
        "action": "create",
        "method": "POST",
        "url": "/api/system/user/create",
        "ip": "192.168.1.1",
        "request": "{\"userName\":\"test\",\"password\":\"******\"}",
        "response": "{\"code\":200,\"message\":\"创建成功\"}",
        "status": 1,
        "errorMsg": "",
        "duration": 50,
        "createdAt": "2023-01-01T12:00:00Z"
      }
    ],
    "page": 1,
    "pageSize": 20,
    "totalPages": 5
  }
}
```

### 2. 查询登录日志

#### 请求信息

- **URL**: `/api/audit/login`
- **方法**: `POST`
- **权限要求**: 需要管理员权限

#### 请求参数

| 参数名    | 类型     | 必填 | 描述                 |
|-----------|----------|------|----------------------|
| userName  | string   | 否   | 用户名（支持模糊查询）|
| startTime | datetime | 否   | 开始时间             |
| endTime   | datetime | 否   | 结束时间             |
| page      | int      | 否   | 页码（默认为1）      |
| pageSize  | int      | 否   | 每页记录数（默认20） |

#### 请求示例

```json
{
  "userName": "admin",
  "page": 1,
  "pageSize": 20
}
```

#### 响应参数

| 参数名     | 类型   | 描述               |
|------------|--------|-------------------|
| code       | int    | 状态码，200表示成功 |
| message    | string | 响应消息           |
| data       | object | 响应数据           |
| data.total | int    | 总记录数           |
| data.list  | array  | 登录日志列表       |
| data.page  | int    | 当前页码           |
| data.pageSize | int  | 每页记录数         |
| data.totalPages | int | 总页数           |

#### 响应示例

```json
{
  "code": 200,
  "message": "请求成功",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "userId": 1,
        "userName": "admin",
        "ip": "192.168.1.1",
        "location": "内网IP",
        "browser": "Chrome",
        "os": "Windows 10",
        "status": 1,
        "msg": "登录成功",
        "createdAt": "2023-01-01T12:00:00Z"
      }
    ],
    "page": 1,
    "pageSize": 20,
    "totalPages": 5
  }
}
```

### 3. 查询ES审计日志

#### 请求信息

- **URL**: `/api/audit/es`
- **方法**: `POST`
- **权限说明**: 管理员可以查看所有用户的记录，普通用户只能查看自己的记录

#### 请求参数

| 参数名      | 类型     | 必填 | 描述                 |
|-------------|----------|------|----------------------|
| project     | string   | 否   | 项目名称             |
| viewName    | string   | 否   | 视图名称             |
| userName    | string   | 否   | 用户名（支持模糊查询）|
| keyword     | string   | 否   | 关键词（搜索关键词或索引模式）|
| startTime   | datetime | 否   | 开始时间             |
| endTime     | datetime | 否   | 结束时间             |
| currentPage | int      | 否   | 页码（默认为1）      |
| pageSize    | int      | 否   | 每页记录数（默认10） |

#### 请求示例

```json
{
  "project": "日志平台",
  "viewName": "业务日志",
  "userName": "admin",
  "keyword": "error",
  "startTime": "2023-01-01 00:00:00",
  "endTime": "2023-01-31 23:59:59",
  "currentPage": 1,
  "pageSize": 10
}
```

#### 响应参数

| 参数名     | 类型   | 描述               |
|------------|--------|-------------------|
| code       | int    | 状态码，200表示成功 |
| message    | string | 响应消息           |
| data       | object | 响应数据           |
| data.total | int    | 总记录数           |
| data.list  | array  | ES查询日志列表     |

#### 响应示例

```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "userId": 1,
        "username": "admin",
        "clientIp": "192.168.1.1",
        "project": "日志平台",
        "viewName": "业务日志",
        "qIndexPattern": "logs-*",
        "qTimeField": "@timestamp",
        "qStartTime": "2023-01-01T00:00:00Z",
        "qEndTime": "2023-01-01T23:59:59Z",
        "qKeyword": "error",
        "responseCode": 200,
        "queryTimeMs": 125,
        "docCount": 56,
        "createdAt": "2023-01-01T10:15:30Z"
      }
    ]
  }
}
```

## 日志记录机制

### 操作日志记录

操作日志通过中间件自动记录，主要记录以下信息：

1. 用户信息：包括用户ID和用户名
2. 操作信息：包括操作模块、操作类型、请求方法和URL
3. 请求信息：包括请求参数（过滤敏感信息如密码）
4. 响应信息：包括响应结果
5. 其他信息：包括操作IP、执行时长、操作状态等

### 登录日志记录

登录日志在用户登录时记录，包括登录成功和失败的情况，主要记录以下信息：

1. 用户信息：包括用户ID和用户名
2. 登录信息：包括登录IP、登录地点、浏览器类型、操作系统
3. 状态信息：包括登录状态（成功/失败）和提示消息

### ES查询日志记录

ES查询日志在用户执行Elasticsearch查询操作时自动记录，主要记录以下信息：

1. 用户信息：包括用户ID、用户名和客户端IP
2. 查询信息：包括项目名称、视图名称、索引模式、时间字段、时间范围和搜索关键词
3. 结果信息：包括响应状态码、查询耗时和文档数量
4. 错误信息：记录查询失败时的错误信息

## 日志查询权限

1. **操作日志和登录日志**：只有具有管理员权限的用户才能查询这些审计日志，确保系统安全和用户隐私。
2. **ES查询日志**：管理员可以查看所有用户的ES查询记录，而普通用户只能查看自己的查询记录。
