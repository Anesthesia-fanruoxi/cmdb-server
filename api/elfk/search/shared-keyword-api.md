# ELFK 共享查询关键词接口文档

## 概述

用于保存和管理 ELFK 查询的共享关键词，方便团队成员复用常用的查询条件。

---

## 1. 创建共享关键词

**接口**: `POST /api/elfk/keyword/create`

**请求头**:
```
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "project": "项目名称",
  "category": "分类名称",
  "view_id": 1,
  "keyword": "查询关键词内容",
  "remark": "备注说明（可选）"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| project | string | 是 | 项目名称 |
| category | string | 是 | 分类名称 |
| view_id | int | 是 | 视图ID |
| keyword | string | 是 | 查询关键词 |
| remark | string | 否 | 备注说明 |

**成功响应**:
```json
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "id": 1,
    "project": "项目名称",
    "category": "分类名称",
    "view_id": 1,
    "keyword": "查询关键词内容",
    "remark": "备注说明",
    "creator": "admin",
    "created_at": "2025-01-01T10:00:00+08:00"
  }
}
```

---

## 2. 查询共享关键词列表（分页）

**接口**: `GET /api/elfk/keyword/list`

**请求头**:
```
Authorization: Bearer {token}
```


**请求参数（Query）**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| project | string | 否 | 项目名称（模糊搜索） |
| category | string | 否 | 分类名称（模糊搜索） |
| view_id | int | 否 | 视图ID（精确匹配） |
| search | string | 否 | 混合搜索（匹配关键词、备注、添加人） |

**请求示例**:
```
GET /api/elfk/keyword/list?page=1&project=cmdb&search=error
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
        "category": "应用日志",
        "view_id": 1,
        "keyword": "error AND timeout",
        "remark": "超时错误查询",
        "creator": "admin",
        "created_at": "2025-01-01T10:00:00+08:00"
      }
    ],
    "total": 100,
    "page": 1,
    "size": 100
  }
}
```

---

## 3. 更新共享关键词

**接口**: `POST /api/elfk/keyword/update`

**请求头**:
```
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "id": 1,
  "project": "项目名称",
  "category": "分类名称",
  "view_id": 1,
  "keyword": "新的查询关键词",
  "remark": "新的备注说明"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 记录ID |
| project | string | 是 | 项目名称 |
| category | string | 是 | 分类名称 |
| view_id | int | 是 | 视图ID |
| keyword | string | 是 | 查询关键词 |
| remark | string | 否 | 备注说明 |

**成功响应**:
```json
{
  "code": 200,
  "message": "更新成功",
  "data": {
    "id": 1,
    "project": "项目名称",
    "category": "分类名称",
    "view_id": 1,
    "keyword": "新的查询关键词",
    "remark": "新的备注说明",
    "creator": "admin",
    "created_at": "2025-01-01T10:00:00+08:00"
  }
}
```

---

## 4. 删除共享关键词

**接口**: `POST /api/elfk/keyword/delete?id={id}`

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
POST /api/elfk/keyword/delete?id=1
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
