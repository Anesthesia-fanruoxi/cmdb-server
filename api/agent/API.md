# CMDB Agent 接口文档

Base URL: `http://{agent_ip}:{port}`

---

## 通用响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

---

## 一、插件管理

### 1. 安装插件

**POST** `/api/plugins/install`

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | 插件名称 |
| version | string | ✅ | 版本号 |
| category | string | ✅ | 类型：`container` 或 `binary` |
| image | string | container必填 | 镜像地址 |
| download_url | string | binary必填 | 二进制下载地址 |
| command | string | ❌ | 启动命令（binary类型可从中提取端口） |
| port | int | ✅ | 服务端口（1-65535） |
| config | object | ❌ | 环境变量/配置参数 |
| config_file | string | ❌ | 配置文件内容（binary类型） |
| parameters.container_port | int | ❌ | 容器内服务端口，默认与 port 相同 |

**container 示例：**
```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "category": "container",
  "image": "registry.example.com/my-plugin:1.0.0",
  "port": 8080,
  "config": {
    "DB_HOST": "127.0.0.1",
    "DB_PORT": "3306"
  },
  "parameters": {
    "container_port": 8080
  }
}
```

**binary 示例：**
```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "category": "binary",
  "download_url": "http://example.com/my-plugin-1.0.0.tar.gz",
  "command": "./my-plugin -port 9090",
  "port": 9090,
  "config": {
    "ENV": "production"
  }
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "name": "my-plugin",
    "version": "1.0.0",
    "category": "container",
    "image": "registry.example.com/my-plugin:1.0.0",
    "container_id": "abc123",
    "port": 8080,
    "status": "running"
  }
}
```

---

### 2. 插件版本升级

**PUT** `/api/plugins/upgrade`

> 只升级版本，不触发配置变动。升级过程：停止旧服务 → 拉取/下载新版本 → 用原配置启动。

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | 插件名称 |
| version | string | ✅ | 新版本号 |
| image | string | container必填 | 新镜像完整地址 |
| download_url | string | ❌ | 新下载地址（binary类型，不填则用旧记录） |

**container 示例：**
```json
{
  "name": "my-plugin",
  "version": "1.2.0",
  "image": "registry.example.com/my-plugin:1.2.0"
}
```

**binary 示例：**
```json
{
  "name": "my-plugin",
  "version": "1.2.0",
  "download_url": "http://example.com/my-plugin-1.2.0.tar.gz"
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "name": "my-plugin",
    "old_version": "1.0.0",
    "new_version": "1.2.0",
    "image": "registry.example.com/my-plugin:1.2.0",
    "container_id": "def456",
    "port": 8080,
    "status": "running",
    "message": "容器版本升级成功"
  }
}
```

---

### 3. 插件配置更新

**PUT** `/api/plugins/update`

> 在原配置基础上做增量变更，支持同时 upsert 多个 key 和删除多个 key，不影响版本。

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | 插件名称 |
| config_set | object | ❌ | 新增或修改的配置项（upsert） |
| config_delete | array | ❌ | 要删除的配置 key 列表 |

> `config_set` 和 `config_delete` 不能同时为空。

**示例（同时修改和删除）：**
```json
{
  "name": "my-plugin",
  "config_set": {
    "DB_HOST": "10.0.0.1",
    "DB_PORT": "5432"
  },
  "config_delete": ["OLD_KEY", "DEPRECATED_ENV"]
}
```

**仅新增/修改：**
```json
{
  "name": "my-plugin",
  "config_set": {
    "NEW_KEY": "value"
  }
}
```

**仅删除：**
```json
{
  "name": "my-plugin",
  "config_delete": ["OLD_KEY"]
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "name": "my-plugin",
    "version": "1.0.0",
    "config": { "DB_HOST": "10.0.0.1", "DB_PORT": "5432" },
    "status": "running",
    "message": "配置更新成功，服务已重启"
  }
}
```

---

### 4. 插件操作（启动/停止/重启）

**PUT** `/api/plugins/operate?name={name}&action={action}`

**Query 参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| name | ✅ | 插件名称 |
| action | ✅ | 操作类型：`start` / `stop` / `restart` |

**响应：**
```json
{
  "code": 0,
  "data": {
    "name": "my-plugin",
    "action": "restart",
    "result": "服务 cmdb-my-plugin 操作成功，当前状态: running",
    "message": "插件 restart 操作成功"
  }
}
```

---

### 5. 卸载插件

**DELETE** `/api/plugins/uninstall?name={name}`

**Query 参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| name | ✅ | 插件名称 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "name": "my-plugin",
    "category": "container",
    "message": "插件 my-plugin 卸载成功"
  }
}
```

---

### 6. 插件列表

**GET** `/api/plugins/list`

**Query 参数（均可选）：**

| 参数 | 说明 |
|------|------|
| category | 按类型过滤：`container` / `binary` |
| status | 按状态过滤：`running` / `stopped` |
| name | 按名称模糊搜索 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "total": 2,
    "agent_version": "1.0.0",
    "eip": "1.2.3.4",
    "plugins": [
      {
        "name": "my-plugin",
        "version": "1.0.0",
        "category": "container",
        "status": "running",
        "port": 8080,
        "uptime": "2d3h",
        "installed_at": "2025-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

### 7. 插件详情

**GET** `/api/plugins/detail?name={name}`

**Query 参数：**

| 参数 | 必填 | 说明 |
|------|------|------|
| name | ✅ | 插件名称 |

**响应：**
```json
{
  "code": 0,
  "data": {
    "name": "my-plugin",
    "version": "1.0.0",
    "category": "container",
    "status": "running",
    "port": 8080,
    "container_port": 8080,
    "uptime": "2d3h",
    "config": { "DB_HOST": "127.0.0.1" },
    "installed_at": "2025-01-01T00:00:00Z"
  }
}
```

---

## 二、日志查看（WebSocket）

### 1. 插件实时日志

**WS** `ws://{agent_ip}:{port}/ws/plugins/logs?name={name}`

- 固定返回最近 1000 行，并持续推送新日志
- container 类型使用 `docker logs -f`
- binary 类型读取插件目录下的 `{name}.log` 文件

---

### 2. CICD 任务日志

**WS** `ws://{agent_ip}:{port}/ws/cicd/logs?data={encrypted}`

`data` 为加密参数，解密后结构：

```json
{
  "taskId": "task-123",
  "stepType": "deployService"
}
```

**stepType 支持值：**
`console` / `pullOnline` / `tagImages` / `pushLocal` / `checkImage` / `deployService` / `checkService` / `trafficSwitching` / `cleanupOldVersion`

---

## 三、插件代理

### 转发请求到插件

**ANY** `/proxy/{plugin-name}/{real-path}`

- 请求体需加密：`{"data": "encrypted_base64"}`
- 响应体同样加密返回：`{"data": "encrypted_base64"}`
- 以下路径为明文直通（不加解密）：`/update`、`/health`、`/ping`、`/api/upload`

**示例：**
```
POST /proxy/my-plugin/search
→ 转发到 http://localhost:8080/search
```

---

## 四、健康检查

**GET** `/health`

**响应：**
```json
{
  "code": 0,
  "data": { "status": "ok" }
}
```

---

## 错误码

| HTTP 状态码 | 说明 |
|-------------|------|
| 400 | 参数错误 |
| 404 | 插件不存在 |
| 405 | 请求方法不允许 |
| 500 | 服务内部错误 |
| 502 | 代理转发失败 |
