# SQL 共享查询接口文档

## 概述

用于保存和管理 SQL 共享查询语句，方便团队成员复用常用的查询。

---

## 1. 创建共享查询

**接口**: `POST /api/sql/shared/create`

**请求头**:
```
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "project": "项目名称",
  "db_name": "数据库名称",
  "query": "SELECT * FROM table WHERE id = 1",
  "remark": "备注说明（可选）"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| project | string | 是 | 项目名称 |
| db_name | string | 是 | 数据库名称 |
| query | string | 是 | 查询语句 |
| remark | string | 否 | 备注说明 |

**成功响应**:
```json
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "id": 1,
    "project": "项目名称",
    "db_name": "数据库名称",
    "query": "SELECT * FROM table WHERE id = 1",
    "remark": "备注说明",
    "creator": "admin",
    "created_at": "2025-01-01T10:00:00+08:00"
  }
}
```

---

## 2. 查询共享查询列表（最多100条）

**接口**: `GET /api/sql/shared/list`

**请求头**:
```
Authorization: Bearer {token}
```

**请求参数（Query）**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| project | string | 否 | 项目名称（模糊搜索） |
| db_name | string | 否 | 数据库名称（模糊搜索） |
| search | string | 否 | 混合搜索（匹配备注、添加人） |


**请求示例**:
```
GET /api/sql/shared/list?project=cmdb&search=admin
```

**成功响应**:
```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "list": [
      {
        "id": 1,
        "project": "cmdb",
        "db_name": "test_db",
        "query": "SELECT * FROM users WHERE status = 1",
        "remark": "查询活跃用户",
        "creator": "admin",
        "created_at": "2025-01-01T10:00:00+08:00"
      }
    ],
    "total": 100
  }
}
```

---

## 3. 更新共享查询

**接口**: `POST /api/sql/shared/update`

**请求头**:
```
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "id": 1,
  "project": "项目名称",
  "db_name": "数据库名称",
  "query": "新的查询语句",
  "remark": "新的备注说明"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 记录ID |
| project | string | 是 | 项目名称 |
| db_name | string | 是 | 数据库名称 |
| query | string | 是 | 查询语句 |
| remark | string | 否 | 备注说明 |

**成功响应**:
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": 1,
    "project": "项目名称",
    "db_name": "数据库名称",
    "query": "新的查询语句",
    "remark": "新的备注说明",
    "creator": "admin",
    "created_at": "2025-01-01T10:00:00+08:00"
  }
}
```

---

## 4. 删除共享查询

**接口**: `POST /api/sql/shared/delete?id={id}`

**请求头**:
```
Authorization: Bearer {token}
```

**请求参数（Query）**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 记录ID |

**请求示例**:
```
POST /api/sql/shared/delete?id=1
```

**成功响应**:
```json
{
  "code": 200,
  "message": "删除成功",
  "data": null
}
```

---

## 错误响应

所有接口错误响应格式：
```json
{
  "code": 400,
  "message": "错误信息"
}
```

**常见错误码**:
| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未登录或token失效 |
| 403 | 无权限 |
| 404 | 记录不存在 |
| 500 | 服务器内部错误 |
