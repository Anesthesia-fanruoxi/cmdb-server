# SQL审批流程接口文档

## 1. 创建审批流程

```
POST /api/sql/process/create
```

### 请求参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| dept_id | BIGINT | 是 | 部门ID |
| apply_id | BIGINT | 是 | 审批人ID |
| executor_id | BIGINT | 是 | 执行人ID |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": ""
}
```

### 错误码
| 错误码 | 描述 |
|--------|------|
| 400 | 参数错误 |
| 500 | 服务器内部错误 |

## 2. 获取审批流程列表

```
GET /api/sql/process/list
```


### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "dept_id": "生产环境",
        "apply_id": ["user1", "user2"],
        "executor_id": ["user3"]
      }
    ],
    "total": 1
  }
}
```

### 错误码
| 错误码 | 描述 |
|--------|------|
| 500 | 服务器内部错误 |

## 3. 获取审批流程详情

```
GET /api/sql/process/detail
```

### 请求参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | int | 是 | 流程ID |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "department": "生产环境",
    "approvers": ["user1", "user2"],
    "executors": ["user3"]
  }
}
```

### 错误码
| 错误码 | 描述 |
|--------|------|
| 400 | 参数错误 |
| 404 | 流程不存在 |
| 500 | 服务器内部错误 |

## 4. 更新审批流程

```
PUT /api/sql/process/update
```

### 请求参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | int | 是 | 流程ID |
| department | string | 否 | 部门名称 |
| approvers | array | 否 | 审核人列表 |
| executors | array | 否 | 执行人列表 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "生产环境SQL审批",
    "description": "更新后的描述"
  }
}
```

### 错误码
| 错误码 | 描述 |
|--------|------|
| 400 | 参数错误 |
| 404 | 流程不存在 |
| 500 | 服务器内部错误 |

## 5. 删除审批流程

```
DELETE /api/sql/process/delete
```

### 请求参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | int | 是 | 流程ID |

### 响应示例

```json
{
  "code": 200,
  "message": "success"
}
```

### 错误码
| 错误码 | 描述 |
|--------|------|
| 400 | 参数错误 |
| 404 | 流程不存在 |
| 500 | 服务器内部错误 |