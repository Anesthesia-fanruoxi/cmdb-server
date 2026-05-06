# 文档管理接口文档

## 1. 文档管理接口

### 1.1 获取文档列表
获取文档列表信息,不包含文档内容。

- 请求路径: `/api/knowledge/doc/list`
- 请求方法: `GET`
- 请求示例: `/api/knowledge/doc/list?page=1&page_size=10`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | int | 否 | 页码,默认1 |
| page_size | int | 否 | 每页条数,默认10 |

- 响应结果:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "total": 100,
        "list": [
            {
                "id": 1,
                "title": "文档标题",
                "category": 1,
                "user_id": 1,
                "user_name": "张三",
                "created_at": "2024-03-20 10:00:00",
                "updated_at": "2024-03-20 10:00:00"
            }
        ],
        "page": {
            "current": 1,
            "size": 10,
            "total": 100,
            "pages": 10
        }
    }
}
```

### 1.2 获取文档详情
获取单个文档的详细信息,包含文档内容。

- 请求路径: `/api/knowledge/doc/detail`
- 请求方法: `GET`
- 请求示例: `/api/knowledge/doc/detail?id=1`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | uint | 是 | 文档ID |

- 响应结果:
```json
{
    "code": 200,
    "message": "获取成功", 
    "data": {
        "id": 1,
        "title": "文档标题",
        "content": "文档内容",
        "category": 1,
        "status": 1,
        "user_id": 1,
        "user_name": "张三",
        "version": 1,
        "created_at": "2024-03-20 10:00:00",
        "updated_at": "2024-03-20 10:00:00"
    }
}
```

### 1.3 创建文档

- 请求路径: `/api/knowledge/doc/create`
- 请求方法: `POST`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| title | string | 是 | 文档标题,2-200字符 |
| content | string | 是 | 文档内容 |
| category | uint | 是 | 分类ID |

- 请求示例:
```json
{
    "title": "文档标题",
    "content": "文档内容",
    "category": 1
}
```

- 响应结果:
```json
{
    "code": 200,
    "message": "创建成功",
    "data": {
        "doc_id": 1
    }
}
```

### 1.4 更新文档

- 请求路径: `/api/knowledge/doc/update`
- 请求方法: `PUT`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | uint | 是 | 文档ID(URL参数) |
| title | string | 是 | 文档标题,2-200字符 |
| content | string | 是 | 文档内容 |
| category | uint | 是 | 分类ID |

- 请求示例:
```json
{
    "title": "更新后的标题",
    "content": "更新后的内容",
    "category": 1
}
```

- 响应结果:
```json
{
    "code": 200,
    "message": "更新成功",
    "data": null
}
```

### 1.5 删除文档

- 请求路径: `/api/knowledge/doc/delete`
- 请求方法: `DELETE`
- 请求示例: `/api/knowledge/doc/delete?id=1`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | uint | 是 | 文档ID(URL参数) |

- 响应结果:
```json
{
    "code": 200,
    "message": "删除成功",
    "data": null
}
```

## 2. 文档历史版本接口

### 2.1 获取历史版本列表

- 请求路径: `/api/knowledge/doc/history/list`
- 请求方法: `GET`
- 请求示例: `/api/knowledge/doc/history/list?doc_id=1`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| doc_id | int64 | 是 | 文档ID |

- 响应结果:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": [
        {
            "id": 1,
            "title": "文档标题",
            "content": "文档内容",
            "version": 1,
            "user_id": 1,
            "created_at": "2024-03-20 10:00:00"
        }
    ]
}
```

### 2.2 获取历史版本详情

- 请求路径: `/api/knowledge/doc/history/detail`
- 请求方法: `GET`
- 请求示例: `/api/knowledge/doc/history/detail?history_id=1`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| history_id | int64 | 是 | 历史版本ID |

- 响应结果:
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": 1,
        "doc_id": 1,
        "title": "文档标题",
        "content": "文档内容",
        "version": 1,
        "user_id": 1,
        "created_at": "2024-03-20 10:00:00"
    }
}
```

### 2.3 恢复历史版本

- 请求路径: `/api/knowledge/doc/history/restore`
- 请求方法: `POST`
- 请求头:
```
Content-Type: application/json
Authorization: Bearer <token>
```
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| history_id | int64 | 是 | 历史版本ID |

- 请求示例:
```json
{
    "history_id": 1
}
```

- 响应结果:
```json
{
    "code": 200,
    "message": "恢复成功",
    "data": null
}
```

## 注意事项

1. 所有接口都需要在请求头中携带 token 进行身份认证
2. 文档列表接口不返回文档内容,以减少数据传输量
3. 更新文档时会自动保存历史版本
4. 删除文档会将文档移入回收站,不会真正删除数据
5. 文档版本号会随着更新自动递增

