# 应用在线更新 API 文档

## 概述

用于 CMDB Desktop 应用的在线更新功能，支持 Windows 和 macOS（Intel/Apple Silicon）平台。

---

## 配置

在 `.env` 中配置：

```bash
# 方式1：使用默认路径（推荐）
# 更新接口为 VITE_API_BASE_URL + /app/version
VITE_API_BASE_URL=https://cmdb.hzbxhd.com/api

# 方式2：自定义更新接口地址
VITE_UPDATE_URL=https://your-server.com/api/app/version
```

---

## 版本检查接口

**接口**: `GET /api/app/version`

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "version": "0.2.0",
    "release_date": "2025-01-15",
    "changelog": "1. 新增 SQL 共享查询功能\n2. 优化日志搜索性能\n3. 修复若干已知问题",
    "mandatory": false,
    "windows": {
      "url": "https://your-cdn.com/releases/cmdb-desktop-0.2.0-x64.msi",
      "size": 52428800,
      "sha256": "abc123..."
    },
    "macos_intel": {
      "url": "https://your-cdn.com/releases/cmdb-desktop-0.2.0-x64.dmg",
      "size": 61865984,
      "sha256": "def456..."
    },
    "macos_arm": {
      "url": "https://your-cdn.com/releases/cmdb-desktop-0.2.0-aarch64.dmg",
      "size": 58720256,
      "sha256": "ghi789..."
    }
  }
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| version | string | 是 | 最新版本号，如 "0.2.0" |
| release_date | string | 是 | 发布日期，如 "2025-01-15" |
| changelog | string | 是 | 更新日志，支持换行符 `\n` |
| mandatory | boolean | 是 | 是否强制更新（true 时用户无法跳过） |
| windows | object | 否 | Windows x64 MSI 安装包 |
| macos_intel | object | 否 | macOS Intel (x64) DMG 安装包 |
| macos_arm | object | 否 | macOS Apple Silicon (M1/M2/M3) DMG 安装包 |

**平台资源字段**:

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 下载地址（建议使用 CDN） |
| size | number | 文件大小（字节） |
| sha256 | string | 文件 SHA256 校验值 |

---

## 更新内容示例

changelog 字段支持多行文本，建议格式：

```
1. 新增功能A
2. 新增功能B
3. 优化XXX性能
4. 修复XXX问题
```

---

## 更新流程

### Windows
1. 客户端定时检查版本接口（默认每 6 小时）
2. 发现新版本后，弹出更新通知
3. 用户点击下载，下载 MSI 安装包
4. 下载完成后，用户点击安装
5. 调用 `msiexec /i xxx.msi /passive` 静默安装
6. 应用自动退出，安装程序接管

### macOS
1. 检查和下载流程与 Windows 相同
2. 客户端自动检测 CPU 架构（Intel x64 或 Apple Silicon arm64）
3. 根据架构选择对应的 DMG 安装包下载
4. 下载完成后，打开 DMG 文件
5. 用户手动将应用拖拽到 Applications 文件夹
6. 需要用户手动重启应用

---

## 前端集成

在 `App.tsx` 中引入更新通知组件并启动自动检查：

```tsx
import { useEffect } from 'react';
import UpdateNotification from './components/UpdateNotification';
import { startAutoCheck } from './services/updater';

function App() {
  useEffect(() => {
    // 启动自动检查（每6小时）
    startAutoCheck(6);
  }, []);

  return (
    <>
      {/* 其他内容 */}
      <UpdateNotification />
    </>
  );
}
```

---

## 手动检查更新

```typescript
import { checkUpdate } from '@/services/updater';

const handleCheckUpdate = async () => {
  try {
    const info = await checkUpdate();
    if (info) {
      console.log('发现新版本:', info.version);
    } else {
      console.log('已是最新版本');
    }
  } catch (err) {
    console.error('检查更新失败:', err);
  }
};
```
