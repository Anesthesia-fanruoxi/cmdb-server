# 任务管理接口文档

## 1. 任务基础操作

### 1.1 获取任务列表
- 请求路径：`/api/job/task/list`
- 请求方法：GET
- 权限要求：`job:task:r`
- 请求参数：
  ```
  ?page=1&page_size=10&name=&task_key=&status=
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
          "name": "数据备份",
          "task_key": "backup_data",
          "cron": "0 0 2 * * ?",
          "description": "每天凌晨2点进行数据备份",
          "status": 1,
          "creator": "admin",
          "created_at": "2024-01-01 12:00:00",
          "updated_at": "2024-01-01 12:00:00"
        }
      ]
    }
  }
  ```

### 1.2 创建任务
- 请求路径：`/api/job/task/create`
- 请求方法：POST
- 权限要求：`job:task:w`
- 请求参数：
  ```json
  {
    "name": "数据备份",           // 任务名称
    "task_key": "backup_data",   // 任务标识符，唯一
    "cron": "0 0 2 * * ?",      // cron表达式
    "description": "每天凌晨2点进行数据备份" // 任务描述
  }
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "创建成功",
    "data": {
      "id": 1,
      "name": "数据备份",
      "task_key": "backup_data",
      "cron": "0 0 2 * * ?",
      "description": "每天凌晨2点进行数据备份",
      "status": 0,
      "creator": "admin",
      "created_at": "2024-01-01 12:00:00",
      "updated_at": "2024-01-01 12:00:00"
    }
  }
  ```

### 1.3 更新任务
- 请求路径：`/api/job/task/update`
- 请求方法：PUT
- 权限要求：`job:task:w`
- 请求参数：
  ```json
  {
    "id": 1,                    // 任务ID
    "name": "数据备份",           // 任务名称，可选
    "cron": "0 0 2 * * ?",      // cron表达式，可选
    "description": "每天凌晨2点进行数据备份", // 任务描述，可选
    "status": 1                 // 状态：0-停止，1-运行，可选
  }
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "更新成功",
    "data": {
      "id": 1,
      "name": "数据备份",
      "task_key": "backup_data",
      "cron": "0 0 2 * * ?",
      "description": "每天凌晨2点进行数据备份",
      "status": 1,
      "creator": "admin",
      "created_at": "2024-01-01 12:00:00",
      "updated_at": "2024-01-01 12:30:00"
    }
  }
  ```

### 1.4 删除任务
- 请求路径：`/api/job/task/delete`
- 请求方法：DELETE
- 权限要求：`job:task:w`
- 请求参数：
  ```
  ?id=1  // 任务ID，URL参数
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "删除成功",
    "data": null
  }
  ```

### 1.5 获取任务详情
- 请求路径：`/api/job/task/detail`
- 请求方法：GET
- 权限要求：`job:task:r`
- 请求参数：
  ```
  ?id=1  // 任务ID，URL参数
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "id": 1,
      "name": "数据备份",
      "task_key": "backup_data",
      "cron": "0 0 2 * * ?",
      "description": "每天凌晨2点进行数据备份",
      "status": 1,
      "creator": "admin",
      "created_at": "2024-01-01 12:00:00",
      "updated_at": "2024-01-01 12:30:00",
      "projects": [
        {
          "project": "cmdb"
        },
        {
          "project": "deploy"
        }
      ]
    }
  }
  ```

### 1.6 验证cron表达式
- 请求路径：`/api/job/task/cron`
- 请求方法：POST
- 权限要求：`job:task:r`
- 请求参数：
  ```json
  {
    "cron": "0 0 2 * * ?"  // cron表达式
  }
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "验证完成",
    "data": {
      "is_valid": true,
      "message": "",
      "next_run": [
        "2024-01-02 02:00:00",
        "2024-01-03 02:00:00",
        "2024-01-04 02:00:00",
        "2024-01-05 02:00:00",
        "2024-01-06 02:00:00"
      ]
    }
  }
  ```
  
  无效表达式的响应示例：
  ```json
  {
    "code": 200,
    "msg": "验证完成",
    "data": {
      "is_valid": false,
      "message": "无效的cron表达式: expected exactly 5 fields, found 4: [0 0 * *]",
      "next_run": []
    }
  }
  ```

## 2. 任务绑定操作

### 2.1 绑定任务
- 请求路径：`/api/job/task/bind`
- 请求方法：POST
- 权限要求：`job:task:w`
- 请求参数：
  ```json
  {
    "job_id": 1,                    // 任务ID
    "projects": ["cmdb", "deploy"]  // 项目标识列表
  }
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "绑定成功",
    "data": null
  }
  ```

### 2.2 获取任务绑定详情
- 请求路径：`/api/job/task/bind/detail`
- 请求方法：GET
- 权限要求：`job:task:r`
- 请求参数：
  ```
  ?job_id=1  // 任务ID
  ```
- 响应示例：
  ```json
  {
    "code": 200,
    "msg": "获取成功",
    "data": {
      "job_id": 1,
      "projects": ["cmdb", "deploy"]  // 已绑定的项目标识列表
    }
  }
  ```

## 状态码说明
- 200: 成功
- 400: 请求参数错误
- 401: 未授权
- 403: 权限不足
- 404: 资源不存在
- 500: 服务器内部错误
