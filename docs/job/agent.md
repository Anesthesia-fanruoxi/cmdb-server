# Agent管理接口文档

## 1. Agent注册与心跳

### 1.1 Agent注册
- 请求路径：`/api/job/agent/register`
- 请求方法：POST
- 权限要求：无需认证
- 请求参数：
  ```json
  {
    "agent_id": "agent-uuid-xxxx",  // Agent唯一标识
    "hostname": "web-server-01",    // 主机名
    "ip": "192.168.1.100",          // IP地址
    "project": "cmdb",              // 项目标识
    "version": "1.0.0",             // Agent版本
    "os_info": "CentOS 7.9 x86_64"  // 操作系统信息
  }
  ```
- 请求示例：
  ```bash
  curl -X POST http://localhost:8080/api/job/agent/register \
    -H "Content-Type: application/json" \
    -d '{
      "agent_id": "agent-uuid-xxxx",
      "hostname": "web-server-01",
      "ip": "192.168.1.100",
      "project": "cmdb",
      "version": "1.0.0",
      "os_info": "CentOS 7.9 x86_64"
    }'
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "注册成功",
    "data": {
      "agent_id": "agent-uuid-xxxx",
      "status": 0  // 0:待审核
    }
  }
  ```

### 1.2 Agent心跳
- 请求路径：`/api/job/agent/heartbeat`
- 请求方法：POST
- 权限要求：无需认证
- 请求参数：
  ```json
  {
    "agent_id": "agent-uuid-xxxx"  // Agent唯一标识
  }
  ```
- 请求示例：
  ```bash
  curl -X POST http://localhost:8080/api/job/agent/heartbeat \
    -H "Content-Type: application/json" \
    -d '{
      "agent_id": "agent-uuid-xxxx"
    }'
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "心跳成功",
    "data": {
      "status": 1  // 1:已授权
    }
  }
  ```
- 错误响应示例：
  ```json
  {
    "code": 404,
    "msg": "Agent不存在或已过期",
    "data": null
  }
  ```

## 2. 任务获取与结果上报

### 2.1 获取任务列表
- 请求路径：`/api/job/agent/task`
- 请求方法：GET
- 权限要求：无需认证，但需要Agent已授权
- 请求参数：
  ```
  ?agent_id=agent-uuid-xxxx&project=cmdb
  ```
  | 参数名 | 类型 | 必填 | 说明 |
  |-------|------|------|------|
  | agent_id | string | 是 | Agent唯一标识 |
  | project | string | 是 | 项目标识 |

- 请求示例：
  ```bash
  curl "http://localhost:8080/api/job/agent/task?agent_id=agent-uuid-xxxx&project=cmdb"
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "tasks": [
        {
          "job_id": 1,
          "name": "数据备份",
          "description": "每天凌晨2点进行数据备份",
          "task_key": "backup_data",
          "project": "cmdb",
          "status": 1,
          "cron": "0 0 2 * * ?",
          "created_at": "2024-01-01 12:00:00",
          "next_run_time": "2024-01-02 02:00:00"
        },
        {
          "job_id": 2,
          "name": "日志清理",
          "description": "每周日凌晨3点清理过期日志",
          "task_key": "clean_logs",
          "project": "cmdb",
          "status": 1,
          "cron": "0 0 3 ? * SUN",
          "created_at": "2024-01-01 12:00:00",
          "next_run_time": "2024-01-07 03:00:00"
        }
      ]
    }
  }
  ```
- 错误响应示例：
  ```json
  {
    "code": 403,
    "msg": "Agent未授权或已禁用",
    "data": null
  }
  ```

### 2.2 上报执行结果
- 请求路径：`/api/job/agent/report`
- 请求方法：POST
- 权限要求：无需认证，但需要Agent已授权
- 请求参数：
  ```json
  {
    "agent_id": "agent-uuid-xxxx",     // Agent唯一标识
    "job_id": 1,                       // 任务ID
    "project": "cmdb",                 // 项目标识
    "task_name": "数据备份任务",        // 任务名称
    "task_key": "backup_data",         // 任务标识符
    "exec_log": "开始备份数据...\n备份完成，共处理1000条记录", // 执行日志
    "exec_status": 1,                  // 执行状态 0:失败 1:成功 2:未执行
    "result": "备份成功，文件保存在/backup/data-20240101.zip", // 执行结果
    "error_msg": "",                   // 错误信息
    "start_time": "2024-01-01 02:00:00", // 开始时间
    "end_time": "2024-01-01 02:01:00",   // 结束时间
    "duration": 60                     // 执行时长(秒)
  }
  ```
- 请求示例：
  ```bash
  curl -X POST http://localhost:8080/api/job/agent/report \
    -H "Content-Type: application/json" \
    -d '{
      "agent_id": "agent-uuid-xxxx",
      "job_id": 1,
      "project": "cmdb",
      "task_name": "数据备份任务",
      "task_key": "backup_data",
      "exec_log": "开始备份数据...\n备份完成，共处理1000条记录",
      "exec_status": 1,
      "result": "备份成功，文件保存在/backup/data-20240101.zip",
      "error_msg": "",
      "start_time": "2024-01-01 02:00:00",
      "end_time": "2024-01-01 02:01:00",
      "duration": 60
    }'
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "上报成功",
    "data": null
  }
  ```
- 错误响应示例：
  ```json
  {
    "code": 403,
    "msg": "Agent未授权或已禁用",
    "data": null
  }
  ```

## 3. Agent管理接口

### 3.1 获取Agent列表
- 请求路径：`/api/job/agent/list`
- 请求方法：GET
- 权限要求：`job:agent:r`
- 请求参数：
  ```
  ?project=&status=
  ```
  | 参数名 | 类型 | 必填 | 说明 |
  |-------|------|------|------|
  | project | string | 否 | 项目标识，用于过滤 |
  | status | int | 否 | 状态：0-待审核，1-已授权，2-已禁用 |

- 请求示例：
  ```bash
  # 获取所有Agent
  curl "http://localhost:8080/api/job/agent/list" \
    -H "Authorization: Bearer {token}"
    
  # 获取特定项目的Agent
  curl "http://localhost:8080/api/job/agent/list?project=cmdb" \
    -H "Authorization: Bearer {token}"
    
  # 获取待审核的Agent
  curl "http://localhost:8080/api/job/agent/list?status=0" \
    -H "Authorization: Bearer {token}"
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "total": 2,
      "list": [
        {
          "agent_id": "agent-uuid-xxxx",
          "hostname": "web-server-01",
          "ip": "192.168.1.100",
          "project": "cmdb",
          "version": "1.0.0",
          "os_info": "CentOS 7.9 x86_64",
          "status": 1,
          "last_heartbeat": "2024-01-01 12:30:00"
        },
        {
          "agent_id": "agent-uuid-yyyy",
          "hostname": "db-server-01",
          "ip": "192.168.1.101",
          "project": "cmdb",
          "version": "1.0.0",
          "os_info": "Ubuntu 20.04 x86_64",
          "status": 0,
          "last_heartbeat": "2024-01-01 12:25:00"
        }
      ]
    }
  }
  ```

### 3.2 Agent授权
- 请求路径：`/api/job/agent/authorize`
- 请求方法：POST
- 权限要求：`job:agent:w`
- 请求参数：
  ```json
  {
    "agent_id": "agent-uuid-xxxx",  // Agent唯一标识
    "status": 1                     // 状态 1:已授权 2:已禁用
  }
  ```
- 请求示例：
  ```bash
  # 授权Agent
  curl -X POST http://localhost:8080/api/job/agent/authorize \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer {token}" \
    -d '{
      "agent_id": "agent-uuid-xxxx",
      "status": 1
    }'
    
  # 禁用Agent
  curl -X POST http://localhost:8080/api/job/agent/authorize \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer {token}" \
    -d '{
      "agent_id": "agent-uuid-xxxx",
      "status": 2
    }'
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "Agent已授权",
    "data": null
  }
  ```
- 错误响应示例：
  ```json
  {
    "code": 404,
    "msg": "Agent不存在或已过期",
    "data": null
  }
  ```

### 3.3 获取Agent详情
- 请求路径：`/api/job/agent/detail`
- 请求方法：GET
- 权限要求：`job:agent:r`
- 请求参数：
  ```
  ?agent_id=agent-uuid-xxxx
  ```
  | 参数名 | 类型 | 必填 | 说明 |
  |-------|------|------|------|
  | agent_id | string | 是 | Agent唯一标识 |

- 请求示例：
  ```bash
  curl "http://localhost:8080/api/job/agent/detail?agent_id=agent-uuid-xxxx" \
    -H "Authorization: Bearer {token}"
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "agent_id": "agent-uuid-xxxx",
      "hostname": "web-server-01",
      "ip": "192.168.1.100",
      "project": "cmdb",
      "version": "1.0.0",
      "os_info": "CentOS 7.9 x86_64",
      "status": 1,
      "last_heartbeat": "2024-01-01 12:30:00",
      "tasks": [
        {
          "job_id": 1,
          "name": "数据备份",
          "description": "每天凌晨2点进行数据备份",
          "task_key": "backup_data",
          "project": "cmdb",
          "status": 1,
          "cron": "0 0 2 * * ?",
          "created_at": "2024-01-01 12:00:00",
          "next_run_time": "2024-01-02 02:00:00"
        }
      ],
      "task_count": 1
    }
  }
  ```
- 错误响应示例：
  ```json
  {
    "code": 404,
    "msg": "Agent不存在或已过期",
    "data": null
  }
  ```

## 4. 状态码说明

| 状态码 | 说明 |
|-------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 权限不足或Agent未授权 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 5. Agent状态说明

| 状态码 | 状态名称 | 说明 |
|-------|---------|------|
| 0 | 待审核 | Agent已注册但尚未被授权 |
| 1 | 已授权 | Agent已被授权，可以获取和执行任务 |
| 2 | 已禁用 | Agent已被禁用，不能获取和执行任务 |

## 6. Agent工作流程

1. **注册流程**：
   - Agent首次启动时，生成唯一ID（可以使用UUID）
   - 调用注册接口，提交主机信息
   - 服务端记录Agent信息，初始状态为"待审核"

2. **授权流程**：
   - 管理员在CMDB系统中查看待审核的Agent列表
   - 确认Agent信息无误后，将其状态更改为"已授权"
   - 也可以将可疑或不再使用的Agent设置为"已禁用"

3. **任务获取流程**：
   - Agent定期发送心跳，更新自己的在线状态
   - Agent调用任务获取接口，提供自己的agent_id和project
   - 服务端验证Agent的授权状态，只有已授权的Agent才能获取任务
   - 返回该项目下需要执行的任务列表

4. **任务执行流程**：
   - Agent获取到任务后，根据任务标识符执行相应的操作
   - 执行完成后，调用上报接口提交执行结果
   - 服务端记录执行结果，用于后续查询和统计

## 7. 常见问题

### 7.1 Agent注册后为什么需要授权？
授权机制可以防止未经授权的客户端获取敏感任务，提高系统安全性。管理员可以审核Agent的来源和信息，确保只有合法的Agent才能执行任务。

### 7.2 Agent心跳间隔应该设置多久？
建议设置为1-5分钟，过短会增加服务器负担，过长则可能导致Agent状态更新不及时。

### 7.3 如何处理Agent离线的情况？
当Agent超过10分钟未发送心跳时，可以认为其已离线。系统应记录最后心跳时间，并在管理界面中标记离线状态，便于管理员排查问题。

### 7.4 如何确保Agent执行任务的安全性？
1. 使用HTTPS加密通信
2. 实施严格的授权机制
3. 限制Agent只能执行特定项目的任务
4. 记录详细的执行日志，便于审计
5. 定期更新Agent版本，修复安全漏洞
