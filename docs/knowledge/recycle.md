# 回收站接口文档

## 1. 回收站列表接口

### 1.1 获取回收站列表

- 请求路径: `/api/knowledge/recycle/list`
- 请求方法: `GET`
- 请求示例: `/api/knowledge/recycle/list?type=doc&page=1&page_size=10`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | string | 是 | 文档类型(doc:公共文档 personal:个人文档) |
| page | int | 否 | 页码,默认1 |
| page_size | int | 否 | 每页条数,默认10 |

- 响应结果(公共文档):
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "total": 100,
        "list": [
            {
                "id": 1,
                "doc_id": 1,
                "title": "文档标题",
                "category": 1,
                "project": "项目名称",
                "deleter_name": "张三",
                "deleted_at": "2024-03-20 10:00:00"
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

- 响应结果(个人文档):
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "total": 100,
        "list": [
            {
                "id": 1,
                "doc_id": 1,
                "title": "文档标题",
                "user_name": "张三",
                "deleted_at": "2024-03-20 10:00:00"
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

## 2. 回收站操作接口

### 2.1 恢复文档

- 请求路径: `/api/knowledge/recycle/refresh`
- 请求方法: `POST`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int64 | 是 | 回收站文档ID |
| type | string | 是 | 文档类型(doc:公共文档 personal:个人文档) |

- 请求示例:
```json
{
    "id": 1,
    "type": "doc"
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

### 2.2 批量恢复文档

- 请求路径: `/api/knowledge/recycle/batch/refresh`
- 请求方法: `POST`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| ids | []int64 | 是 | 回收站文档ID列表 |
| type | string | 是 | 文档类型(doc:公共文档 personal:个人文档) |

- 请求示例:
```json
{
    "ids": [1, 2, 3],
    "type": "doc"
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

### 2.3 彻底删除文档

- 请求路径: `/api/knowledge/recycle/remove`
- 请求方法: `DELETE`
- 请求示例: `/api/knowledge/recycle/remove?id=1&type=doc`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int64 | 是 | 回收站文档ID |
| type | string | 是 | 文档类型(doc:公共文档 personal:个人文档) |

- 响应结果:
```json
{
    "code": 200,
    "message": "删除成功",
    "data": null
}
```

### 2.4 批量彻底删除文档

- 请求路径: `/api/knowledge/recycle/batch/remove`
- 请求方法: `POST`
- 请求参数:

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| ids | []int64 | 是 | 回收站文档ID列表 |
| type | string | 是 | 文档类型(doc:公共文档 personal:个人文档) |

- 请求示例:
```json
{
    "ids": [1, 2, 3],
    "type": "doc"
}
```

- 响应结果:
```json
{
    "code": 200,
    "message": "删除成功",
    "data": null
}
```

## 注意事项

1. 所有接口都需要在请求头中携带 token 进行身份认证
2. 文档恢复后会保持原有的属性(标题、内容等)
3. 彻底删除后的文档无法恢复
4. type 参数用于区分操作的是公共文档还是个人文档
5. 批量操作时建议每次不超过100条数据
