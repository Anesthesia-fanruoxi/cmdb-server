# 任务执行记录接口文档

## 1. 执行记录列表

### 1.1 获取执行记录列表
- 请求路径：`/api/job/exec/list`
- 请求方法：GET
- 权限要求：`job:task:r`
- 请求参数：
  ```
  ?page=1&page_size=10&job_id=&project=&task_name=&exec_status=
  ```
  | 参数名 | 类型 | 必填 | 说明 |
  |-------|------|------|------|
  | page | int | 否 | 页码，默认1 |
  | page_size | int | 否 | 每页数量，默认10 |
  | job_id | int | 否 | 任务ID |
  | project | string | 否 | 项目标识 |
  | task_name | string | 否 | 任务名称，支持模糊查询 |
  | exec_status | int | 否 | 执行状态：0-失败，1-成功，2-未执行 |

- 请求示例：
  ```
  # 基本请求
  GET /api/job/exec/list?page=1&page_size=10
  
  # 按项目和状态过滤
  GET /api/job/exec/list?project=cmdb&exec_status=1
  
  # 按任务ID和任务名称过滤
  GET /api/job/exec/list?job_id=1&task_name=备份
  
  # 组合过滤
  GET /api/job/exec/list?project=cmdb&exec_status=0&task_name=同步&page=2&page_size=20
  ```

- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "total": 100,
      "list": [
        {
          "id": 1,
          "job_id": 1,
          "job_name": "数据备份",
          "agent_id": "agent-uuid-xxxx",
          "project": "cmdb",
          "task_key": "backup_data",
          "task_name": "数据备份任务",
          "exec_status": 1,
          "start_time": "2024-01-01 02:00:00",
          "end_time": "2024-01-01 02:01:00",
          "duration": 60,
          "created_at": "2024-01-01 02:01:01"
        },
        {
          "id": 2,
          "job_id": 2,
          "job_name": "日志清理",
          "agent_id": "agent-uuid-yyyy",
          "project": "cmdb",
          "task_key": "clean_logs",
          "task_name": "日志清理任务",
          "exec_status": 1,
          "start_time": "2024-01-01 03:00:00",
          "end_time": "2024-01-01 03:00:30",
          "duration": 30,
          "created_at": "2024-01-01 03:00:31"
        }
      ]
    }
  }
  ```

> 注意：列表接口只返回基本信息，不包含执行日志、执行结果和错误信息等大文本字段，以提高接口性能。

## 2. 执行记录详情

### 2.1 获取执行记录详情
- 请求路径：`/api/job/exec/detail`
- 请求方法：GET
- 权限要求：`job:task:r`
- 请求参数：
  ```
  ?id=1  // 执行记录ID
  ```
  | 参数名 | 类型 | 必填 | 说明 |
  |-------|------|------|------|
  | id | int | 是 | 执行记录ID |

- 请求示例：
  ```
  # 获取ID为1的执行记录详情
  GET /api/job/exec/detail?id=1
  
  # 获取ID为42的执行记录详情
  GET /api/job/exec/detail?id=42
  ```

- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "id": 1,
      "job_id": 1,
      "job_name": "数据备份",
      "agent_id": "agent-uuid-xxxx",
      "project": "cmdb",
      "task_key": "backup_data",
      "task_name": "数据备份任务",
      "exec_status": 1,
      "exec_log": "开始备份数据...\n连接数据库成功\n开始备份表: users\n备份表: users 完成，共 1000 条记录\n开始备份表: orders\n备份表: orders 完成，共 5000 条记录\n备份完成，共处理6000条记录\n压缩备份文件...\n备份文件压缩完成",
      "result": "备份成功，文件保存在/backup/data-20240101.zip，大小: 15MB",
      "error_msg": "",
      "start_time": "2024-01-01 02:00:00",
      "end_time": "2024-01-01 02:01:00",
      "duration": 60,
      "created_at": "2024-01-01 02:01:01"
    }
  }
  ```

- 错误响应示例：
  ```json
  {
    "code": 404,
    "msg": "执行记录不存在",
    "data": null
  }
  ```

> 注意：详情接口返回完整信息，包括执行日志、执行结果和错误信息等大文本字段。

## 3. 执行状态说明

| 状态码 | 状态名称 | 说明 |
|-------|---------|------|
| 0 | 失败 | 任务执行失败，通常会有错误信息 |
| 1 | 成功 | 任务执行成功 |
| 2 | 未执行 | 任务未能执行，可能是因为Agent问题或其他原因 |

## 4. 字段说明

| 字段名 | 类型 | 说明 |
|-------|------|------|
| id | int | 执行记录ID |
| job_id | int | 任务ID |
| job_name | string | 任务名称 |
| agent_id | string | 执行Agent的ID |
| project | string | 项目标识 |
| task_key | string | 任务标识符 |
| task_name | string | 任务名称 |
| exec_log | string | 执行日志，记录任务执行的详细过程 |
| exec_status | int | 执行状态：0-失败，1-成功，2-未执行 |
| result | string | 执行结果，通常是任务执行的最终输出 |
| error_msg | string | 错误信息，当任务执行失败时的错误详情 |
| start_time | string | 开始时间，格式：yyyy-MM-dd HH:mm:ss |
| end_time | string | 结束时间，格式：yyyy-MM-dd HH:mm:ss |
| duration | int | 执行时长(秒) |
| created_at | string | 创建时间，格式：yyyy-MM-dd HH:mm:ss |

## 5. 常见问题

### 5.1 如何查询特定时间段内的执行记录？
目前API不直接支持按时间范围查询，但您可以获取较大数量的记录后在客户端进行过滤。

### 5.2 执行日志过大时如何处理？
对于非常大的执行日志，建议在前端实现分段显示或折叠显示功能，以提高用户体验。

### 5.3 如何处理执行状态为"未执行"的记录？
状态为"未执行"的记录通常表示Agent在接收到任务后，由于某些原因未能执行。建议检查Agent的状态和日志，确认是否存在连接问题或权限问题。
