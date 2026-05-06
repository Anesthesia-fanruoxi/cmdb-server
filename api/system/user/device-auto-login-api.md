# 设备自动登录接口文档

## 概述

本文档描述了 CMDB 桌面客户端的设备绑定和自动登录功能的后端接口设计。

### 安全机制

采用 **挑战-响应（Challenge-Response）** 机制，确保：
1. 设备密钥（device_secret）永远不会在网络上传输
2. 每次登录请求都是唯一的，无法被重放攻击
3. 即使攻击者知道用户名和设备码，也无法暴力破解

### 流程图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           设备绑定流程                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  用户已登录（有 Token）                                                   │
│       │                                                                 │
│       ▼                                                                 │
│  点击"绑定设备"，输入 TOTP 验证码                                         │
│       │                                                                 │
│       ▼                                                                 │
│  POST /system/user/device/bind                                          │
│  请求: { user_name, machine_id, totp_code }                             │
│  Header: token                                                          │
│       │                                                                 │
│       ▼                                                                 │
│  后端验证 TOTP，生成 device_secret                                       │
│  存储: user_name + machine_id + device_secret                           │
│       │                                                                 │
│       ▼                                                                 │
│  返回 device_secret 给客户端                                             │
│  客户端加密存储到本地                                                     │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                           自动登录流程                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  客户端启动，检测到本地有设备凭证                                          │
│       │                                                                 │
│       ▼                                                                 │
│  步骤1: 请求挑战码                                                       │
│  POST /system/user/device/challenge                                     │
│  请求: { user_name, machine_id }                                        │
│       │                                                                 │
│       ▼                                                                 │
│  后端生成随机 challenge（30秒有效）                                       │
│  存储: challenge -> { user_name, machine_id, expires_at }               │
│       │                                                                 │
│       ▼                                                                 │
│  返回: { challenge, expires_at }                                        │
│       │                                                                 │
│       ▼                                                                 │
│  步骤2: 客户端生成签名                                                   │
│  message = "{challenge}:{timestamp}"                                    │
│  signature = HMAC-SHA256(message, device_secret)                        │
│       │                                                                 │
│       ▼                                                                 │
│  步骤3: 发送登录请求                                                     │
│  POST /system/user/device/login                                         │
│  请求: { user_name, machine_id, challenge, timestamp, signature }       │
│       │                                                                 │
│       ▼                                                                 │
│  后端验证:                                                               │
│  1. challenge 是否存在且未过期                                           │
│  2. challenge 对应的 user_name 和 machine_id 是否匹配                    │
│  3. timestamp 是否在合理范围内（±60秒）                                   │
│  4. 使用存储的 device_secret 重新计算签名并比对                           │
│       │                                                                 │
│       ▼                                                                 │
│  验证通过，返回 Token                                                    │
│  删除已使用的 challenge（一次性）                                         │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 数据库表设计

### 设备绑定表 `sys_user_device`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| user_id | bigint | 用户ID |
| user_name | varchar(64) | 用户名 |
| machine_id | varchar(128) | 设备硬件指纹 |
| device_secret | varchar(128) | 设备密钥（64位随机字符串） |
| device_name | varchar(64) | 设备名称（可选，用于展示） |
| last_login_at | datetime | 最后登录时间 |
| created_at | datetime | 绑定时间 |
| status | tinyint | 状态：1-正常，0-禁用 |

**唯一索引**: `(user_name)` - 一个用户只能绑定一个设备

### 挑战码缓存（Redis）

```
Key: device:challenge:{challenge}
Value: JSON { "user_name": "xxx", "machine_id": "xxx", "expires_at": 1703750400 }
TTL: 30秒
```

---

## 接口详情

### 1. 绑定设备

**接口**: `POST /system/user/device/bind`

**说明**: 用户登录后，绑定当前设备。需要双因子验证确保是本人操作。

**请求头**:
```
Authorization: Bearer {用户登录token}
```

**请求体**:
```json
{
  "machine_id": "a1b2c3d4e5f6...",
  "totp_code": "123456"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| machine_id | string | 是 | 设备硬件指纹（SHA256哈希值，64位十六进制） |
| totp_code | string | 是 | 6位TOTP验证码 |

**成功响应**:
```json
{
  "code": 200,
  "message": "绑定成功",
  "data": {
    "device_secret": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2"
  }
}
```

**失败响应**:
```json
{
  "code": 400,
  "message": "TOTP验证码错误"
}
```

**后端处理逻辑**:
```python
def bind_device(token, machine_id, totp_code):
    # 1. 验证 token，获取当前用户
    current_user = verify_token(token)
    
    # 2. 验证 TOTP
    if not verify_totp(current_user.totp_secret, totp_code):
        return error("TOTP验证码错误")
    
    # 3. 生成设备密钥（64位随机字符串）
    device_secret = secrets.token_hex(32)  # 64位十六进制
    
    # 4. 保存或更新绑定记录（用户只能绑定一个设备，重新绑定会覆盖）
    db.execute("""
        INSERT INTO sys_user_device (user_id, user_name, machine_id, device_secret, created_at)
        VALUES (?, ?, ?, ?, NOW())
        ON DUPLICATE KEY UPDATE machine_id = ?, device_secret = ?, created_at = NOW()
    """, current_user.id, current_user.user_name, machine_id, device_secret, machine_id, device_secret)
    
    # 5. 返回设备密钥
    return success({"device_secret": device_secret})
```

---

### 4. 解绑设备

**接口**: `POST /system/user/device/unbind`

**说明**: 用户解绑当前设备，需要双因子验证确保是本人操作。

**请求头**:
```
Authorization: Bearer {用户登录token}
```

**请求体**:
```json
{
  "totp_code": "123456"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| totp_code | string | 是 | 6位TOTP验证码 |

**成功响应**:
```json
{
  "code": 200,
  "message": "解绑成功",
  "data": null
}
```

**失败响应**:
```json
{
  "code": 400,
  "message": "TOTP验证码错误"
}
```

```json
{
  "code": 400,
  "message": "该设备未绑定"
}
```

**后端处理逻辑**:
```python
def unbind_device(token, totp_code):
    # 1. 验证 token，获取当前用户
    current_user = verify_token(token)
    
    # 2. 验证 TOTP
    if not verify_totp(current_user.totp_secret, totp_code):
        return error("TOTP验证码错误")
    
    # 3. 获取请求中的 machine_id（从客户端获取当前设备指纹）
    # 注意：实际实现中 machine_id 应该从请求体或客户端传递
    machine_id = get_machine_id_from_request()
    
    # 4. 检查设备是否已绑定
    device = db.query(
        "SELECT * FROM sys_user_device WHERE user_name = ? AND machine_id = ?",
        current_user.user_name, machine_id
    )
    if not device:
        return error("该设备未绑定")
    
    # 5. 删除绑定记录
    db.execute(
        "DELETE FROM sys_user_device WHERE user_name = ? AND machine_id = ?",
        current_user.user_name, machine_id
    )
    
    # 6. 返回成功
    return success(None, "解绑成功")
```

---

### 2. 获取挑战码

**接口**: `POST /system/user/device/challenge`

**说明**: 自动登录第一步，获取一次性挑战码。

**请求体**:
```json
{
  "user_name": "admin",
  "machine_id": "a1b2c3d4e5f6..."
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_name | string | 是 | 用户名 |
| machine_id | string | 是 | 设备硬件指纹 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "challenge": "x7y8z9a0b1c2d3e4f5g6h7i8j9k0l1m2",
    "expires_at": 1703750430
  }
}
```

**失败响应**:
```json
{
  "code": 400,
  "message": "设备未绑定"
}
```

```json
{
  "code": 429,
  "message": "请求过于频繁，请稍后再试"
}
```

**后端处理逻辑**:
```python
def get_challenge(user_name, machine_id):
    # 1. 检查设备是否已绑定
    device = db.query(
        "SELECT * FROM sys_user_device WHERE user_name = ? AND machine_id = ? AND status = 1",
        user_name, machine_id
    )
    if not device:
        return error("设备未绑定")
    
    # 2. 频率限制（同一设备每分钟最多5次）
    rate_key = f"device:challenge:rate:{machine_id}"
    if redis.incr(rate_key) > 5:
        return error("请求过于频繁，请稍后再试", code=429)
    redis.expire(rate_key, 60)
    
    # 3. 生成挑战码（32位随机字符串）
    challenge = secrets.token_hex(16)
    expires_at = int(time.time()) + 30  # 30秒后过期
    
    # 4. 存储到 Redis
    challenge_key = f"device:challenge:{challenge}"
    redis.setex(challenge_key, 30, json.dumps({
        "user_name": user_name,
        "machine_id": machine_id,
        "expires_at": expires_at
    }))
    
    # 5. 返回挑战码
    return success({
        "challenge": challenge,
        "expires_at": expires_at
    })
```

---

### 3. 设备登录

**接口**: `POST /system/user/device/login`

**说明**: 自动登录第二步，使用签名验证身份。

**请求体**:
```json
{
  "user_name": "admin",
  "machine_id": "a1b2c3d4e5f6...",
  "challenge": "x7y8z9a0b1c2d3e4f5g6h7i8j9k0l1m2",
  "timestamp": 1703750400,
  "signature": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_name | string | 是 | 用户名 |
| machine_id | string | 是 | 设备硬件指纹 |
| challenge | string | 是 | 从上一步获取的挑战码 |
| timestamp | int | 是 | 客户端当前时间戳（秒） |
| signature | string | 是 | HMAC-SHA256签名（64位十六进制） |

**签名算法**:
```
message = "{challenge}:{timestamp}"
signature = HMAC-SHA256(message, device_secret)
```

**成功响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": "1",
    "user_name": "admin"
  }
}
```

**失败响应**:
```json
{
  "code": 400,
  "message": "挑战码无效或已过期"
}
```

```json
{
  "code": 400,
  "message": "签名验证失败"
}
```

```json
{
  "code": 400,
  "message": "时间戳超出有效范围"
}
```

**后端处理逻辑**:
```python
import hmac
import hashlib

def device_login(user_name, machine_id, challenge, timestamp, signature):
    # 1. 验证挑战码
    challenge_key = f"device:challenge:{challenge}"
    challenge_data = redis.get(challenge_key)
    if not challenge_data:
        return error("挑战码无效或已过期")
    
    challenge_info = json.loads(challenge_data)
    
    # 2. 验证挑战码对应的用户和设备
    if challenge_info["user_name"] != user_name or challenge_info["machine_id"] != machine_id:
        return error("挑战码不匹配")
    
    # 3. 验证时间戳（允许±60秒误差）
    current_time = int(time.time())
    if abs(current_time - timestamp) > 60:
        return error("时间戳超出有效范围")
    
    # 4. 获取设备密钥
    device = db.query(
        "SELECT * FROM sys_user_device WHERE user_name = ? AND machine_id = ? AND status = 1",
        user_name, machine_id
    )
    if not device:
        return error("设备未绑定或已禁用")
    
    # 5. 验证签名
    message = f"{challenge}:{timestamp}"
    expected_signature = hmac.new(
        device.device_secret.encode('utf-8'),
        message.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    
    if not hmac.compare_digest(signature, expected_signature):
        # 记录失败次数，防止暴力破解
        fail_key = f"device:login:fail:{machine_id}"
        fail_count = redis.incr(fail_key)
        redis.expire(fail_key, 300)  # 5分钟内
        
        if fail_count >= 5:
            # 禁用设备
            db.execute(
                "UPDATE sys_user_device SET status = 0 WHERE machine_id = ?",
                machine_id
            )
            return error("签名验证失败次数过多，设备已被禁用")
        
        return error("签名验证失败")
    
    # 6. 删除已使用的挑战码（一次性）
    redis.delete(challenge_key)
    
    # 7. 清除失败计数
    redis.delete(f"device:login:fail:{machine_id}")
    
    # 8. 更新最后登录时间
    db.execute(
        "UPDATE sys_user_device SET last_login_at = NOW() WHERE id = ?",
        device.id
    )
    
    # 9. 获取用户信息，生成 Token
    user = db.query("SELECT * FROM sys_user WHERE user_name = ?", user_name)
    token = generate_jwt_token(user)
    
    return success({
        "token": token,
        "user_id": str(user.id),
        "user_name": user.user_name
    })
```

---

## 安全措施总结

### 1. 挑战-响应机制
- 设备密钥永远不在网络传输
- 每次登录使用不同的挑战码
- 挑战码30秒过期，一次性使用

### 2. 频率限制
- 获取挑战码：同一设备每分钟最多5次
- 登录失败：5次失败后禁用设备

### 3. 时间戳验证
- 客户端时间戳必须在服务器时间±60秒内
- 防止重放攻击

### 4. 双因子绑定
- 绑定设备时需要TOTP验证码
- 确保是用户本人操作

### 5. 设备绑定规则
- 一个用户只能绑定一个设备（重新绑定会覆盖旧设备）
- 一个设备可以被多个用户绑定
- 基于硬件指纹识别设备

---

## 客户端实现参考

### Rust 签名生成代码

```rust
use hmac::{Hmac, Mac};
use sha2::Sha256;

type HmacSha256 = Hmac<Sha256>;

fn generate_signature(challenge: &str, timestamp: i64, device_secret: &str) -> String {
    let message = format!("{}:{}", challenge, timestamp);
    let mut mac = HmacSha256::new_from_slice(device_secret.as_bytes())
        .expect("HMAC can take key of any size");
    mac.update(message.as_bytes());
    let result = mac.finalize();
    hex::encode(result.into_bytes())
}
```

### Python 签名验证代码

```python
import hmac
import hashlib

def verify_signature(challenge: str, timestamp: int, signature: str, device_secret: str) -> bool:
    message = f"{challenge}:{timestamp}"
    expected = hmac.new(
        device_secret.encode('utf-8'),
        message.encode('utf-8'),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)
```

### Java 签名验证代码

```java
import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.security.MessageDigest;

public class SignatureUtil {
    public static boolean verifySignature(
        String challenge, 
        long timestamp, 
        String signature, 
        String deviceSecret
    ) {
        try {
            String message = challenge + ":" + timestamp;
            Mac mac = Mac.getInstance("HmacSHA256");
            SecretKeySpec secretKey = new SecretKeySpec(
                deviceSecret.getBytes("UTF-8"), 
                "HmacSHA256"
            );
            mac.init(secretKey);
            byte[] hash = mac.doFinal(message.getBytes("UTF-8"));
            String expected = bytesToHex(hash);
            return MessageDigest.isEqual(
                signature.getBytes(), 
                expected.getBytes()
            );
        } catch (Exception e) {
            return false;
        }
    }
    
    private static String bytesToHex(byte[] bytes) {
        StringBuilder sb = new StringBuilder();
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }
}
```

---

## 测试用例

### 1. 正常绑定流程
```bash
# 1. 用户登录获取 token
curl -X POST http://localhost:8080/system/user/login \
  -H "Content-Type: application/json" \
  -d '{"user_name":"admin","totp_code":"123456"}'

# 2. 绑定设备
curl -X POST http://localhost:8080/system/user/device/bind \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "user_name": "admin",
    "machine_id": "abc123def456",
    "totp_code": "654321"
  }'
```

### 2. 正常自动登录流程
```bash
# 1. 获取挑战码
curl -X POST http://localhost:8080/system/user/device/challenge \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "admin",
    "machine_id": "abc123def456"
  }'

# 2. 设备登录（需要计算签名）
curl -X POST http://localhost:8080/system/user/device/login \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "admin",
    "machine_id": "abc123def456",
    "challenge": "{challenge}",
    "timestamp": 1703750400,
    "signature": "{calculated_signature}"
  }'
```

### 3. 异常测试
- 使用过期的挑战码
- 使用错误的签名
- 使用不匹配的 user_name/machine_id
- 频繁请求挑战码（触发限流）
- 多次签名验证失败（触发设备禁用）
