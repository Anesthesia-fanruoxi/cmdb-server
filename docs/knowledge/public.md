# 公开文档接口文档

## 1. 获取文档列表
### 接口描述
获取公开文档列表

### 请求信息
- 请求路径：`/api/knowledge/public/list`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 请求参数
| 参数名   | 类型   | 必填 | 描述     | 说明 |
|----------|--------|------|----------|------|
| category | string | 否   | 文档分类 | 筛选指定分类的文档 |
| keyword  | string | 否   | 搜索关键词 | 搜索标题、内容、分类 |
| mine     | bool   | 否   | 只看自己 | true=只看自己创建的 |

### 响应示例
```json
{
    "code": 200,
    "message": "获取成功",
    "data": [
        {
            "id": 1,
            "title": "文档标题",
            "category": "文档分类",
            "creator": "创建者",
            "user_name": "最后修改者",
            "created_at": "2024-01-01 12:00:00",
            "updated_at": "2024-01-02 12:00:00"
        }
    ]
}
```

## 2. 获取文档详情
### 接口描述
获取公开文档详情

### 请求信息
- 请求路径：`/api/knowledge/public/detail?id={id}`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 请求参数
| 参数名 | 类型 | 必填 | 描述   | 说明 |
|--------|------|------|--------|------|
| id     | int  | 是   | 文档ID |      |

### 响应示例
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": 1,
        "title": "文档标题",
        "content": "文档内容",
        "category": "文档分类",
        "creator": "创建者",
        "user_name": "最后修改者",
        "created_at": "2024-01-01 12:00:00",
        "updated_at": "2024-01-02 12:00:00"
    }
}
```

## 3. 创建文档
### 接口描述
创建新的公开文档

### 请求信息
- 请求路径：`/api/knowledge/public/create`
- 请求方法：POST
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名   | 类型   | 必填 | 描述     | 验证规则 |
|----------|--------|------|----------|----------|
| title    | string | 是   | 文档标题 | 2-200字符 |
| content  | string | 是   | 文档内容 | 非空 |
| category | string | 是   | 文档分类 | 非空 |

### 响应示例
```json
{
    "code": 200,
    "message": "创建成功",
    "data": {
        "id": 1,
        "title": "文档标题",
        "content": "文档内容",
        "category": "文档分类",
        "creator": "创建者",
        "user_name": "创建者",
        "created_at": "2024-01-01 12:00:00",
        "updated_at": "2024-01-01 12:00:00"
    }
}
```

## 4. 更新文档
### 接口描述
更新公开文档内容

### 请求信息
- 请求路径：`/api/knowledge/public/update`
- 请求方法：PUT
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名   | 类型   | 必填 | 描述     | 验证规则 |
|----------|--------|------|----------|----------|
| id       | int    | 是   | 文档ID   | 非空 |
| title    | string | 是   | 文档标题 | 2-200字符 |
| content  | string | 是   | 文档内容 | 非空 |
| category | string | 是   | 文档分类 | 非空 |

### 响应示例
```json
{
    "code": 200,
    "message": "更新成功",
    "data": {
        "id": 1,
        "title": "更新后的标题",
        "content": "更新后的内容",
        "category": "文档分类",
        "creator": "创建者",
        "user_name": "修改者",
        "created_at": "2024-01-01 12:00:00",
        "updated_at": "2024-01-02 12:00:00"
    }
}
```

## 5. 删除文档
### 接口描述
删除公开文档（移入回收站）

### 请求信息
- 请求路径：`/api/knowledge/public/delete?id={id}`
- 请求方法：DELETE
- 请求头：Authorization: Bearer {token}

### 请求参数
| 参数名 | 类型 | 必填 | 描述   | 说明 |
|--------|------|------|--------|------|
| id     | int  | 是   | 文档ID |      |

### 响应示例
```json
{
    "code": 200,
    "message": "删除成功"
}
```

## 6. 分享文档
### 接口描述
创建文档的分享链接

### 请求信息
- 请求路径：`/api/knowledge/public/share`
- 请求方法：POST
- 请求头：
  - Content-Type: application/json
  - Authorization: Bearer {token}

### 请求参数
| 参数名      | 类型  | 必填 | 描述   | 验证规则 |
|------------|-------|------|--------|----------|
| doc_id     | int64 | 是   | 文档ID | 非空 |
| expired_days| int64 | 是   | 过期天数| >0 |

### 响应示例
```json
{
    "code": 200,
    "message": "分享成功",
    "data": {
        "share_code": "abcd1234",
        "expired_at": "2024-02-01 12:00:00"
    }
}
```

## 7. 匿名访问
### 接口描述
通过分享码匿名访问文档

### 请求信息
- 请求路径：`/api/knowledge/public/anonymous?code={code}`
- 请求方法：GET

### 请求参数
| 参数名 | 类型   | 必填 | 描述   | 说明 |
|--------|--------|------|--------|------|
| code   | string | 是   | 分享码 |      |

### 响应示例
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": 1,
        "title": "文档标题",
        "content": "文档内容",
        "category": "文档分类",
        "creator": "创建者",
        "created_at": "2024-01-01 12:00:00"
    }
}
```

## 8. 关闭分享
### 接口描述
提前关闭文档分享链接

### 请求信息
- 请求路径：`/api/knowledge/public/anonymous/delete?id={id}`
- 请求方法：DELETE
- 请求头：Authorization: Bearer {token}

### 请求参数
| 参数名 | 类型   | 必填 | 描述   | 说明 |
|--------|--------|------|--------|------|
| id     | int64  | 是   | 文档ID |      |

### 响应示例
```json
{
    "code": 200,
    "message": "关闭分享成功"
}
```

## 9. 历史版本列表
### 接口描述
获取文档的历史版本列表

### 请求信息
- 请求路径：`/api/knowledge/public/history/list?doc_id={doc_id}`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 请求参数
| 参数名  | 类型  | 必填 | 描述   | 说明 |
|---------|-------|------|--------|------|
| doc_id  | int64 | 是   | 文档ID |      |

### 响应示例
```json
{
    "code": 200,
    "message": "获取成功",
    "data": [
        {
            "id": 1,
            "title": "文档标题",
            "category": "文档分类",
            "version": 2,
            "user_name": "修改者",
            "created_at": "2024-01-02 12:00:00"
        }
    ]
}
```

## 10. 历史版本详情
### 接口描述
获取历史版本的详细内容

### 请求信息
- 请求路径：`/api/knowledge/public/history/detail?id={id}`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 请求参数
| 参数名 | 类型  | 必填 | 描述       | 说明 |
|--------|-------|------|------------|------|
| id     | int64 | 是   | 历史版本ID |      |

### 响应示例
```json
{
    "code": 200,
    "message": "获取成功",
    "data": {
        "id": 1,
        "doc_id": 1,
        "title": "文档标题",
        "content": "历史版本内容",
        "category": "文档分类",
        "version": 2,
        "user_name": "修改者",
        "created_at": "2024-01-02 12:00:00"
    }
}
```

## 11. 恢复历史版本
### 接口描述
将文档恢复到指定的历史版本

### 请求信息
- 请求路径：`/api/knowledge/public/history/restore`
- 请求方法：POST
- 请求头：Authorization: Bearer {token}

### 请求参数
| 参数名     | 类型  | 必填 | 描述       | 说明 |
| ---------- | ----- | ---- | ---------- | ---- |
| history_id | int64 | 是   | 历史版本ID |      |

### 响应示例
```json
{
    "code": 200,
    "message": "恢复成功",
    "data": {
        "id": 1,
        "title": "文档标题",
        "content": "恢复的内容",
        "category": "文档分类",
        "creator": "创建者",
        "user_name": "恢复者",
        "version": 3,
        "created_at": "2024-01-01 12:00:00",
        "updated_at": "2024-01-03 12:00:00"
    }
}
```

## 12. 分享列表
### 接口描述
获取当前用户的文档分享列表

### 请求信息
- 请求路径：`/api/knowledge/public/share/list`
- 请求方法：GET
- 请求头：Authorization: Bearer {token}

### 响应示例
```json
{
    "code": 200,
    "message": "获取成功",
    "data": [
        {
            "id": 1,
            "doc_id": 1,
            "title": "文档标题",
            "share_code": "abcd1234",
            "share_url": "http://example.com/share?code=abcd1234",
            "creator": "创建者",
            "expired_at": "2024-02-01 12:00:00",
            "created_at": "2024-01-01 12:00:00"
        }
    ]
}
```

## 错误码说明
| 错误码 | 描述                |
|--------|-------------------|
| 400    | 请求参数错误        |
| 401    | 未登录或token失效   |
| 403    | 无权限操作          |
| 404    | 资源不存在          |
| 500    | 服务器内部错误      |

## 注意事项
1. 除了匿名访问接口，其他接口都需要登录并携带token
2. 只有文档创建者可以修改和删除文档
3. 只有文档创建者可以创建和关闭分享链接
4. 分享链接过期后无法访问
5. 文档删除后会进入回收站，可以恢复
6. 每次修改文档会自动保存历史版本
7. 恢复历史版本会生成新的版本号
8. 历史版本按版本号倒序排列
