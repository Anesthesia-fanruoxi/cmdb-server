# DataBI 异步刷新功能说明

## 功能概述

DataBI 模块实现了基于 Redis 的读写分离架构，通过两个独立接口实现异步刷新和 SSE 实时推送，解决了 Hive 查询慢导致前端超时的问题。

## 架构设计

### 两层缓存机制
1. **插件层缓存**：插件启动时将 Hive 表加载到内存
2. **CMDB 层缓存**：将表列表和刷新状态缓存到 Redis

### 读写分离架构
```
refresh 接口（写）          tables 接口（读）
     ↓                          ↓
  后台任务                   SSE 连接
     ↓                          ↓
  写入 Redis  ←────────────  读取 Redis
  (状态+数据)                (轮询状态)
```

### 刷新流程
```
用户点击刷新按钮
  ↓
POST /api/sql/databi/refresh （触发后台任务）
  ↓
同时建立 SSE 连接
  ↓
GET /api/sql/databi/tables （实时读取 Redis）
  ↓
后台任务执行：
  1. 写入状态：refreshing（20%）
  2. 调用插件 loadTable
  3. 写入状态：loading（60%）
  4. 调用插件 tables
  5. 写入表数据到 Redis
  6. 写入状态：success（100%）+ 表数据
  ↓
SSE 实时推送状态变化给前端
```

## API 接口

### 1. 触发刷新任务（写入 Redis）
```
POST /api/sql/databi/refresh?project=xxx
```

**功能**：触发后台异步刷新任务，更新 Redis 中的状态和数据

**响应**：
```json
{
  "code": 200,
  "message": "刷新任务已启动",
  "data": {
    "project": "xxx",
    "status": "started"
  }
}
```

如果已有刷新任务在进行中：
```json
{
  "code": 200,
  "message": "刷新任务已在进行中",
  "data": {
    "project": "xxx",
    "status": "refreshing"
  }
}
```

### 2. SSE 实时推送表列表（读取 Redis）
```
GET /api/sql/databi/tables?project=xxx&token=xxx
```

**功能**：建立 SSE 连接，轮询 Redis 获取刷新状态并实时推送

**参数**：
- `project`（必填）：项目名称
- `token`（必填）：JWT token，通过 URL 参数传递

**SSE 事件类型**：

1. `idle` - 空闲状态，需要刷新
```
event: idle
data: {"status":"idle","message":"请点击刷新按钮"}
```

2. `cached` - 有缓存数据（无刷新任务时）
```
event: cached
data: {"status":"cached","message":"使用缓存数据","tables":{...},"update_time":"2024-01-01T12:00:00Z"}
```

3. `refreshing` - 正在刷新插件缓存
```
event: refreshing
data: {"status":"refreshing","message":"正在刷新插件缓存...","progress":20}
```

4. `loading` - 正在加载表列表
```
event: loading
data: {"status":"loading","message":"正在加载表列表...","progress":60}
```

5. `success` - 刷新成功
```
event: success
data: {"status":"success","message":"刷新完成","progress":100,"tables":{...},"update_time":"2024-01-01T12:00:00Z"}
```

6. `error` - 刷新失败
```
event: error
data: {"status":"error","message":"刷新失败","error":"错误详情"}
```

## 前端集成示例

### Token 认证说明

根据项目现有架构，SSE 接口通过 URL 参数传递 token：

```javascript
const token = localStorage.getItem('token');
const eventSource = new EventSource(
  `/api/sql/databi/tables?project=${project}&token=${token}`
);
```

### 完整流程
```javascript
async function handleRefresh(project) {
  const token = localStorage.getItem('token');
  
  // 1. 触发刷新任务
  const response = await fetch(`/api/sql/databi/refresh?project=${project}`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const result = await response.json();
  if (result.code !== 200) {
    console.error('触发刷新失败:', result.message);
    return;
  }
  
  // 2. 建立 SSE 连接（token 通过 URL 传递）
  const eventSource = new EventSource(
    `/api/sql/databi/tables?project=${project}&token=${token}`
  );
  
  // 3. 监听各种状态事件
  eventSource.addEventListener('idle', (e) => {
    const data = JSON.parse(e.data);
    console.log('空闲状态:', data.message);
  });
  
  eventSource.addEventListener('cached', (e) => {
    const data = JSON.parse(e.data);
    console.log('使用缓存:', data.tables);
    updateTableList(data.tables);
    eventSource.close();
  });
  
  eventSource.addEventListener('refreshing', (e) => {
    const data = JSON.parse(e.data);
    updateProgress(data.progress, data.message);
  });
  
  eventSource.addEventListener('loading', (e) => {
    const data = JSON.parse(e.data);
    updateProgress(data.progress, data.message);
  });
  
  eventSource.addEventListener('success', (e) => {
    const data = JSON.parse(e.data);
    console.log('刷新成功:', data.tables);
    updateTableList(data.tables);
    hideProgress();
    eventSource.close();
  });
  
  eventSource.addEventListener('error', (e) => {
    const data = JSON.parse(e.data);
    console.error('刷新失败:', data.error);
    showError(data.error);
    hideProgress();
    eventSource.close();
  });
  
  // 4. 处理连接错误
  eventSource.onerror = (err) => {
    console.error('SSE 连接错误:', err);
    eventSource.close();
  };
}
```

### Vue 3 完整示例
```vue
<template>
  <div>
    <button @click="handleRefresh" :disabled="isRefreshing">
      {{ isRefreshing ? '刷新中...' : '刷新表列表' }}
    </button>
    
    <div v-if="isRefreshing" class="progress">
      <div class="progress-bar" :style="{ width: progress + '%' }"></div>
      <span>{{ progressMessage }}</span>
    </div>
    
    <div v-if="tables">
      <div v-for="(tableList, dbName) in tables" :key="dbName">
        <h3>{{ dbName }}</h3>
        <ul>
          <li v-for="table in tableList" :key="table">{{ table }}</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';

const project = ref('your_project');
const tables = ref(null);
const isRefreshing = ref(false);
const progress = ref(0);
const progressMessage = ref('');
let eventSource = null;

async function handleRefresh() {
  isRefreshing.value = true;
  progress.value = 0;
  
  const token = localStorage.getItem('token');
  
  try {
    // 1. 触发刷新任务
    const response = await fetch(`/api/sql/databi/refresh?project=${project.value}`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    const result = await response.json();
    if (result.code !== 200) {
      alert('触发刷新失败: ' + result.message);
      isRefreshing.value = false;
      return;
    }
    
    // 2. 建立 SSE 连接（token 通过 URL 传递）
    eventSource = new EventSource(
      `/api/sql/databi/tables?project=${project.value}&token=${token}`
    );
    
    eventSource.addEventListener('idle', (e) => {
      const data = JSON.parse(e.data);
      progressMessage.value = data.message;
    });
    
    eventSource.addEventListener('cached', (e) => {
      const data = JSON.parse(e.data);
      tables.value = data.tables;
      progressMessage.value = '使用缓存数据';
      isRefreshing.value = false;
      eventSource.close();
    });
    
    eventSource.addEventListener('refreshing', (e) => {
      const data = JSON.parse(e.data);
      progress.value = data.progress;
      progressMessage.value = data.message;
    });
    
    eventSource.addEventListener('loading', (e) => {
      const data = JSON.parse(e.data);
      progress.value = data.progress;
      progressMessage.value = data.message;
    });
    
    eventSource.addEventListener('success', (e) => {
      const data = JSON.parse(e.data);
      tables.value = data.tables;
      progress.value = 100;
      progressMessage.value = '刷新完成';
      isRefreshing.value = false;
      eventSource.close();
    });
    
    eventSource.addEventListener('error', (e) => {
      const data = JSON.parse(e.data);
      alert('刷新失败: ' + data.error);
      isRefreshing.value = false;
      eventSource.close();
    });
    
    eventSource.onerror = () => {
      isRefreshing.value = false;
      if (eventSource) {
        eventSource.close();
      }
    };
  } catch (error) {
    alert('请求失败: ' + error.message);
    isRefreshing.value = false;
  }
}
</script>
```

## Redis Key 说明

- `cmdb:databi:tables:{project}` - 表列表缓存（24小时过期）
- `cmdb:databi:refresh:status:{project}` - 刷新状态（10分钟过期）
- `cmdb:databi:refresh:lock:{project}` - 刷新锁（5分钟过期，防止死锁）

## 优势

1. **读写分离**：refresh 接口负责写入，tables 接口负责读取，职责清晰
2. **避免超时**：异步执行，前端不会因为 Hive 查询慢而超时
3. **防止并发**：使用 Redis 锁，同一项目同时只能有一个刷新任务
4. **实时反馈**：通过 SSE 实时推送刷新进度（20% → 60% → 100%）
5. **性能优化**：表列表缓存到 Redis，查询时直接读取缓存
6. **用户体验**：单个刷新按钮，实时显示刷新进度和结果
7. **架构一致**：使用 EventSource + URL token，与项目现有 SSE 接口保持一致
