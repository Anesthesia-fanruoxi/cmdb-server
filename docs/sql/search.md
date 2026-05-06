# MySQL 结构查询接口文档

## 1. 查询数据库列表

获取指定 Agent 下所有可访问的数据库列表。

### 请求

- 请求方法: `POST`
- 请求路径: `/api/sql/search/db`
- Content-Type: `application/json`

### 请求参数

```json
{
  "agent": "bxhd",
  "type": "db",
  "dbName": null,
  "tbName": null
}
```

### 响应

成功响应示例：

```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "databases": [
      "cmdb",
      "information_schema",
      "mysql",
      "performance_schema",
      "sys",
      "test"
    ]
  }
}
```

失败响应示例：

```json
{
    "error": "错误信息"
}
```

## 2. 查询表列表

获取指定数据库中的所有表列表。

### 请求

- 请求方法: `POST`
- 请求路径: `/api/sql/search/tb`
- Content-Type: `application/json`

### 请求参数

```json
{
  "agent": "bxhd",
  "type": "db",
  "dbName": "test",
  "tbName": null
}
```

### 响应

成功响应示例：

```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "tables": [
      "test"
    ]
  }
}
```

失败响应示例：

```json
{
    "error": "错误信息"
}
```

## 3. 查询表结构

获取指定表的详细结构信息。

### 请求

- 请求方法: `POST`
- 请求路径: `/api/sql/search/tb-detail`
- Content-Type: `application/json`

### 请求参数

```json
{
  "agent": "bxhd",
  "type": "db",
  "dbName": "test",
  "tbName": "test"
}
```

### 响应

成功响应示例：

```json
{
  "code": 200,
  "message": "查询成功",
  "data": {
    "group": {
      "default": "NULL",
      "extra": "",
      "field": "group",
      "key": "",
      "null": "YES",
      "type": "text"
    },
    "id": {
      "default": "NULL",
      "extra": "auto_increment",
      "field": "id",
      "key": "PRI",
      "null": "NO",
      "type": "int"
    },
    "name": {
      "default": "NULL",
      "extra": "",
      "field": "name",
      "key": "",
      "null": "YES",
      "type": "text"
    }
  }
}
```

失败响应示例：

```json
{
    "error": "错误信息"
}
```

## 4. 查询数据（分页）

执行SQL查询并获取第一页结果。系统会自动将完整查询结果存储在服务器端，并返回查询ID，用于后续分页请求。

### 请求

- 请求方法: `POST`
- 请求路径: `/api/sql/search/data`
- Content-Type: `application/json`

### 请求参数

```json
{
    "agent": "bxhd",
    "dbName": "test",
    "query": "select * from test"
}
```

### 响应

成功响应示例：

```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "columns": ["id", "name", "group"],
        "db_name": "test",
        "rows": [
            [1, "gggg", "gsgsd"],
            [2, "kkk", null],
            // ...最多20条记录
        ],
        "total": 25,
        "page": 1,
        "page_size": 20,
        "total_pages": 2,
        "query_id": "query_1234567890"
    }
}
```

## 5. 获取分页数据

获取已执行查询的特定页面数据。

### 请求

- 请求方法: `GET`
- 请求路径: `/api/sql/search/page?query_id=query_1234567890&page=2&page_size=20`

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|------|------|
| query_id | string | 是 | 查询ID，从初始查询响应中获取 |
| page | int | 否 | 页码，默认为1 |
| page_size | int | 否 | 每页记录数，默认为20，最大为100 |

### 响应

成功响应示例：

```json
{
    "code": 200,
    "message": "查询成功",
    "data": {
        "columns": ["id", "name", "group"],
        "rows": [
            [21, "dasfasfasfasfas", "546464"],
            [22, "22", "22"],
            [23, null, "23"],
            [24, "24", "24"],
            [25, null, null]
        ],
        "db_name": "test",
        "total": 25,
        "page": 2,
        "page_size": 20,
        "total_pages": 2,
        "query_id": "query_1234567890"
    }
}
```

### 错误响应

```json
{
    "code": 404,
    "message": "查询结果不存在或已过期"
}
```

注意：查询结果在服务器端保存30分钟，超时后将被自动删除。

## 错误码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 安全说明

1. 所有请求和响应数据都经过加密处理
2. 使用 AES-GCM 加密算法
3. 数据传输过程中使用 gzip 压缩
4. 需要配置正确的加密盐才能正常通信

## 注意事项

1. 所有接口都需要正确配置 Agent 和加密盐
2. 查询表结构时需要有对应数据库和表的访问权限
3. 建议在生产环境中使用 HTTPS 协议
4. 接口调用频率建议不要超过每秒 10 次

## 数据加密说明

### 请求数据加密

1. 将原始请求数据转换为 JSON 字符串
2. 使用 gzip 压缩 JSON 字符串
3. 使用 AES-GCM 加密压缩后的数据
4. 将加密后的数据进行 Base64 编码

### 响应数据解密

1. 对接收到的 Base64 编码数据进行解码
2. 使用 AES-GCM 解密数据
3. 使用 gzip 解压数据
4. 解析 JSON 字符串得到原始数据

## 6. WebSocket SQL查询接口

为了提供更安全、更高效的SQL查询体验，系统提供了基于WebSocket的查询接口。该接口支持实时进度反馈、分页数据传输和长连接查询，有效防止大量数据泄露。

### 连接建立

- WebSocket URL: `ws://your-server/api/sql/search/ws`
- 需要在URL中包含有效的JWT令牌，与HTTP接口认证方式相同

### 消息格式

所有WebSocket消息都使用JSON格式，基本结构如下：

```json
{
  "type": "消息类型",
  "query_id": "查询ID（可选）",
  "data": {
    // 消息内容，根据消息类型不同而不同
  }
}
```

### 消息类型

#### 1. 查询请求 (客户端发送)

```json
{
  "type": "query",
  "data": {
    "agent": "bxhd",
    "dbName": "test",
    "query": "SELECT * FROM users LIMIT 1000"
  }
}
```

#### 2. 进度更新 (服务端发送)

```json
{
  "type": "progress",
  "query_id": "query_123456789",
  "data": {
    "percent": 50,
    "message": "正在处理查询结果..."
  }
}
```

#### 3. 获取分页数据 (客户端发送)

```json
{
  "type": "fetch_page",
  "query_id": "query_123456789",
  "data": {
    "page": 2,
    "page_size": 50
  }
}
```

#### 4. 分页数据 (服务端发送)

```json
{
  "type": "page_data",
  "query_id": "query_123456789",
  "data": {
    "page": 2,
    "page_size": 50,
    "total_pages": 20,
    "total_records": 987,
    "columns": ["id", "name", "created_at"],
    "rows": [
      [51, "用户51", "2023-01-01"],
      [52, "用户52", "2023-01-02"],
      // ...更多行数据
    ]
  }
}
```

#### 5. 错误消息 (服务端发送)

```json
{
  "type": "error",
  "query_id": "query_123456789",
  "data": {
    "message": "查询执行失败: 表不存在"
  }
}
```

#### 6. 查询完成 (服务端发送)

```json
{
  "type": "complete",
  "query_id": "query_123456789",
  "data": {
    "total_rows": 987,
    "columns": 3
  }
}
```

### 查询流程

1. **连接建立**：客户端连接WebSocket端点
2. **发送查询**：客户端发送`query`类型消息
3. **接收进度**：服务端发送`progress`类型消息，报告查询进度
4. **查询完成**：服务端发送`complete`类型消息，表示查询已完成
5. **接收数据**：服务端自动发送第一页数据（`page_data`类型消息）
6. **请求更多数据**：客户端发送`fetch_page`类型消息请求其他页的数据
7. **错误处理**：如果查询过程中出现错误，服务端发送`error`类型消息

### 示例代码

#### JavaScript客户端示例

```javascript
// 建立WebSocket连接
const ws = new WebSocket('ws://your-server/api/sql/search/ws?token=your-jwt-token');

// 当前查询ID
let currentQueryId = null;

// 连接建立时的处理
ws.onopen = function(event) {
  console.log('WebSocket连接已建立');
  
  // 发送查询请求
  executeQuery('bxhd', 'test', 'SELECT * FROM users LIMIT 1000');
};

// 接收消息的处理
ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  
  switch(message.type) {
    case 'progress':
      // 更新进度条
      updateProgressBar(message.data.percent, message.data.message);
      break;
      
    case 'page_data':
      // 显示查询结果
      displayQueryResults(message.data.columns, message.data.rows);
      // 更新分页控件
      updatePagination(message.data.page, message.data.total_pages);
      break;
      
    case 'error':
      // 显示错误信息
      showError(message.data.message);
      break;
      
    case 'complete':
      // 查询完成的处理
      console.log(`查询完成，共${message.data.total_rows}条记录`);
      break;
  }
};

// 执行查询
function executeQuery(agent, dbName, query) {
  const message = {
    type: 'query',
    data: {
      agent: agent,
      dbName: dbName,
      query: query
    }
  };
  
  ws.send(JSON.stringify(message));
}

// 获取特定页的数据
function fetchPage(queryId, page, pageSize = 50) {
  const message = {
    type: 'fetch_page',
    query_id: queryId,
    data: {
      page: page,
      page_size: pageSize
    }
  };
  
  ws.send(JSON.stringify(message));
}
```

### 安全说明

1. **认证与授权**：WebSocket连接需要有效的JWT令牌，与HTTP接口使用相同的认证机制
2. **数据分页**：查询结果分页传输，客户端在任何时刻只能访问有限的数据子集
3. **数据存储**：查询结果临时存储在服务器端，设置30分钟过期时间
4. **加密传输**：所有WebSocket通信都应通过HTTPS/WSS进行，确保传输安全
5. **查询限制**：服务器对查询复杂度和结果集大小有限制，防止资源耗尽

### 优势

1. **实时反馈**：查询执行过程中提供实时进度反馈
2. **更高效率**：避免HTTP请求的开销，特别是对于频繁查询
3. **更好安全性**：客户端在任何时刻只能访问有限的数据子集，防止大量数据泄露
4. **更好体验**：提供更流畅的用户体验，特别是对于大型查询

### 注意事项

1. **连接管理**：客户端应妥善管理WebSocket连接，包括断线重连
2. **错误处理**：客户端应处理各种错误情况，包括连接错误和查询错误
3. **资源限制**：服务器对每个用户的WebSocket连接数有限制
4. **兼容性**：确保客户端浏览器支持WebSocket（几乎所有现代浏览器都支持）
