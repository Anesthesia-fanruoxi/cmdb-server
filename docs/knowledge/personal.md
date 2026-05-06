## 3. 个人文档接口

### 3.1 获取个人文档列表

- 请求路径: `/api/knowledge/personal/list`
- 请求方法: `GET`
- 请求示例: `/api/knowledge/personal/list?page=1&page_size=10`
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
                "creator_name": "张三",
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

### 3.2 获取个人文档详情

- 请求路径: `/api/knowledge/personal/detail`
- 请求方法: `GET`
- 请求示例: `/api/knowledge/personal/detail?id=1`
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
        "project": "项目名称",
        "status": 1,
        "creator_name": "张三",
        "user_name": "张三",
        "created_at": "2024-03-20 10:00:00",
        "updated_at": "2024-03-20 10:00:00"
    }
}
```

### 3.3 创建个人文档

- 请求路径: `/api/knowledge/personal/create`
- 请求方法: `POST`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| title | string | 是 | 文档标题,2-200字符 |
| content | string | 是 | 文档内容 ||

- 请求示例:
```json
{
    "title": "文档标题",
    "content": "文档内容"
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

### 3.4 更新个人文档

- 请求路径: `/api/knowledge/personal/update`
- 请求方法: `PUT`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | uint | 是 | 文档ID |
| title | string | 是 | 文档标题,2-200字符 |
| content | string | 是 | 文档内容 |

- 请求示例:
```json
{
    "id": 1,
    "title": "更新后的标题",
    "content": "更新后的内容"
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

### 3.5 删除个人文档

- 请求路径: `/api/knowledge/personal/delete`
- 请求方法: `DELETE`
- 请求示例: `/api/knowledge/personal/delete?id=1`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | uint | 是 | 文档ID |

- 响应结果:
```json
{
    "code": 200,
    "message": "删除成功",
    "data": null
}
```
