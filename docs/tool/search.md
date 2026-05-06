# 日志搜索接口文档

## 1. 日志搜索

### 1.1 搜索日志
- 请求路径：`/api/elfk/search/list`
- 请求方法：POST
- 请求参数：
  ```json
  {
    "index_pattern": "logs-*",   // 索引名称，支持排除模式
    "project": "cmdb",           // 项目标识，必填
    "log_type": "elfk",          // 日志类型，可选值：elfk（默认）或sls
    "start_time": "2025-03-10 00:00:00", // 开始时间，必填，中国时区
    "end_time": "2025-03-11 00:00:00",   // 结束时间，必填，中国时区
    "keyword": "error",          // 搜索关键词，可选。不传时查询时间范围内最新的数据
    "size": 50                   // 返回结果数量，默认50
  }
  ```

### 1.2 获取数据分布统计
- 请求路径：`/api/elfk/search/histograms`
- 请求方法：POST
- 请求参数：与搜索日志接口相同
  ```json
  {
    "index_pattern": "logs-*",   // 索引名称，支持排除模式
    "project": "cmdb",           // 项目标识，必填
    "log_type": "elfk",          // 日志类型，可选值：elfk（默认）或sls
    "start_time": "2025-03-10 00:00:00", // 开始时间，必填，中国时区
    "end_time": "2025-03-11 00:00:00",   // 结束时间，必填，中国时区
    "keyword": "error"           // 搜索关键词，可选
  }
  ```

- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取数据分布成功",
    "data": {
      "total": 145631,           // 总数据量
      "histograms": [            // 时间分布
        {
          "from": 1746692535,    // 开始时间戳
          "to": 1746692550,      // 结束时间戳
          "count": 2274,         // 该时间段内的数据量
          "progress": "Complete" // 查询完成状态
        },
        {
          "from": 1746692550,
          "to": 1746692580,
          "count": 5126,
          "progress": "Complete"
        }
        // ...更多时间段
      ]
    }
  }
  ```

### 关于日志类型 (log_type)
系统支持两种日志查询方式：
1. `elfk`：使用Elasticsearch查询，适用于存储在ES中的日志数据
2. `sls`：使用阿里云SLS（日志服务）查询，适用于存储在阿里云SLS中的日志数据

根据项目配置的不同，请选择对应的日志类型。不同类型的日志查询语法可能略有差异。

索引名称支持以下格式：
- 单个模式：`"logs-*"`
- 包含排除：`"logs-*,-logs-test-*"`

- 关键词语法：
  - 单个关键词：`error`
  - AND 查询：`error and exception`
  - OR 查询：`error or warning`
  - NOT 查询：`error not exception` 或 `error and !exception`
  - 字段查询：`level=error` 或 `status!=200`
  - 复杂组合：`level=error and message=timeout or status!=200`

- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "搜索成功",
    "data": {
      "took": 30,
      "hits": {
        "total": {
          "value": 100,
          "relation": "eq"
        },
        "hits": [
          {
            "_index": "logs-20250310",
            "_id": "xxx",
            "_source": {
              "timestamp": "2025-03-10 10:00:00",
              "level": "error",
              "message": "Connection refused",
              "service": "api-server"
            }
          }
        ]
      }
    }
  }
  ```

### 1.3 获取索引列表
- 请求路径：`/api/elfk/search/indices`
- 请求方法：GET
- 请求参数：
  ```
  ?project=cmdb&log_type=elfk  // project必填，log_type可选(elfk或sls)
  ```

- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "indices": [
        "logs-20250310",
        "logs-20250309",
        "logs-20250308"
      ],
      "fields": [
        {
          "name": "timestamp",
          "type": "date"
        },
        {
          "name": "level",
          "type": "keyword"
        },
        {
          "name": "message",
          "type": "text"
        }
      ]
    }
  }
  ```

### 1.4 获取索引字段结构
- 请求路径：`/api/elfk/search/fields`
- 请求方法：GET
- 请求参数：
  ```
  ?project=cmdb&index=logs&log_type=elfk  // project和index都是必填，log_type可选(elfk或sls)
  ```

- 说明：
  - 对于ELFK模式，接口会自动尝试两种日期格式：
    1. `YYYY.MM.DD` (如：logs-2025.03.10)
    2. `YYYYMMDD` (如：logs-20250310)
  - 对于SLS模式，接口会直接查询SLS的索引名称
  - 如果查询不到索引，则返回404错误

- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "properties": {
        "@timestamp": {
          "type": "date"
        },
        "level": {
          "type": "keyword"
        },
        "message": {
          "type": "text"
        },
        "service": {
          "type": "keyword"
        }
      }
    }
  }
  ```

## 2. 错误码说明

| 状态码 | 说明 |
|-------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 3. 查询语法说明

### 3.1 基础查询
- 单个关键词：直接输入要搜索的内容
  ```
  error
  ```

### 3.2 逻辑运算
- AND 查询：使用 `and` 连接两个条件
  ```
  error and timeout
  ```

- OR 查询：使用 `or` 连接两个条件
  ```
  error or warning
  ```

- NOT 查询：使用 `not` 或 `!` 排除某个条件
  ```
  error not timeout
  error and !exception
  ```

### 3.3 字段查询
- 等于查询：`字段=值`
  ```
  level=error
  service=api-server
  ```

- 不等于查询：`字段!=值`
  ```
  status!=200
  level!=info
  ```

### 3.4 复杂查询示例
1. 查找错误级别且包含超时的日志：
   ```
   level=error and message=timeout
   ```

2. 查找非200状态码或包含错误的日志：
   ```
   status!=200 or message=error
   ```

3. 查找特定服务的错误日志但排除超时：
   ```
   service=api-server and level=error and !timeout
   ```

## 4. 使用建议

### 4.1 时间范围
- 建议将时间范围控制在较小范围内（如1小时内）以获得更好的性能
- 时间格式使用中国时区（如：2025-03-10 00:00:00）

### 4.2 查询优化
- 使用精确的字段查询比全文搜索更快
- 避免使用过于模糊的关键词
- 合理使用逻辑运算符组合查询条件

### 4.3 结果限制
- 默认返回50条结果
- 如需更多结果，可以通过size参数指定，但建议不要超过1000
- 如果需要获取大量数据，建议使用分页或导出功能

## 5. 常见问题

### 5.1 找不到索引
- 检查项目标识是否正确
- 检查日期格式是否匹配
- 确认索引是否已经创建
- 检查log_type是否正确（elfk或sls）

### 5.2 查询超时
- 缩小时间范围
- 优化查询条件
- 避免使用过于复杂的查询语句

### 5.3 结果不准确
- 检查查询语法是否正确
- 确认字段名称是否正确
- 注意关键词的大小写敏感性

- 说明：
  - 当不传递 keyword 参数时，将返回指定时间范围内最新的数据
  - 结果按时间倒序排列，即最新的数据在前
  - 如果设置了 size 参数（如50），则返回最新的50条数据
  - 如果传递了 keyword 参数，则在时间范围内搜索匹配的数据

## 6. 索引视图管理

### 6.1 创建索引视图
- 请求路径：`/api/elfk/search/view/create`
- 请求方法：POST
- 请求参数：
  ```json
  {
    "name": "业务日志视图",        // 视图名称，必填
    "project": "cmdb",           // 项目标识，必填
    "category": "业务",           // 分类，必填
    "log_type": "elfk",          // 日志类型，必填，可选值：elfk或sls
    "index_pattern": "b_info-*", // 索引模式，必填
    "time_field": "@timestamp",  // 时间字段，必填
    "time_format": "yyyy-MM-dd HH:mm:ss", // 时间格式，可选
    "description": "业务相关日志"  // 视图描述，可选
  }
  ```

### 6.2 索引模式语法
索引模式支持以下语法：
- 使用 `*` 作为通配符匹配任意字符
- 使用逗号分隔多个模式
- 使用减号 `-` 前缀表示排除该模式

示例：
1. 单个模式：
   ```
   b_info-*
   ```
   匹配所有以 "b_info-" 开头的索引

2. 多个包含模式：
   ```
   b_info-*,c_info-*
   ```
   匹配所有以 "b_info-" 或 "c_info-" 开头的索引

3. 包含和排除组合：
   ```
   b_info-*,-b_info-test-*
   ```
   匹配所有以 "b_info-" 开头的索引，但排除以 "b_info-test-" 开头的索引

4. 复杂组合：
   ```
   b_info-*,c_info-*,-*-test-*,-*-temp-*
   ```
   匹配以 "b_info-" 或 "c_info-" 开头的索引，但排除包含 "-test-" 或 "-temp-" 的索引

### 6.3 时间字段说明
时间字段用于：
1. 日志搜索时的时间范围过滤
2. 搜索结果的时间排序
3. 时间聚合分析

常见的时间字段：
- `@timestamp`: 标准的 ES 时间字段
- `create_time`: 数据创建时间
- `log_time`: 日志产生时间

注意事项：
1. 时间字段必须是 date 类型
2. 字段必须存在于所有匹配的索引中
3. 建议使用标准的中国时区格式时间字段

### 6.4 获取视图详情
- 请求路径：`/api/elfk/view/detail`
- 请求方法：GET
- 请求参数：
  ```
  ?id=1  // 视图ID，必填
  ```

- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取视图详情成功",
    "data": {
      "id": 1,
      "name": "业务日志视图",
      "project": "cmdb",
      "category": "业务",
      "log_type": "elfk",
      "index_pattern": "b_info-*",
      "time_field": "@timestamp",
      "time_format": "yyyy-MM-dd HH:mm:ss",
      "description": "业务相关日志",
      "update_time": "2025-03-10 10:00:00",
      "all_field": [
        {
          "name": "@timestamp",
          "type": "date"
        },
        {
          "name": "level",
          "type": "keyword"
        }
      ],
      "indices": [
        "b_info-20250310",
        "b_info-20250309"
      ]
    }
  }
  ```
