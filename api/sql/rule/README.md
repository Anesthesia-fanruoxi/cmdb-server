# SQL规则检查系统

## 概述

新的SQL规则检查系统基于完整的SQL解析结果，使用通用的字段检查引擎，配置驱动，易于扩展。

## 文件说明

### 核心文件

- **rulesEngine.go** - 通用规则检查引擎
  - 支持15种检查类型
  - 支持嵌套字段路径访问
  - 自动类型转换

- **rulesCheck.go** - 规则检查API入口
  - `POST /api/sql/rules/check` - 检查SQL规则
  - 支持多条SQL批量检查
  - 返回详细的检查结果（包含完整的SQL分析结果）

### 管理文件

- **rulesList.go** - 规则列表查询
- **rulesUpdate.go** - 规则启用/禁用

## 规则配置

### 规则分类

- **security** - 安全类规则（防止数据丢失、误操作）
- **standard** - 规范类规则（代码规范、最佳实践）

### 检查类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `empty` | 字段为空则失败 | 检查主键是否为空 |
| `not_empty` | 字段不为空则失败 | 禁止自定义字符集 |
| `equals` | 字段等于某值则失败 | 检查布尔值 |
| `in` | 字段在列表中则通过 | 字符集白名单 |
| `not_in` | 字段不在列表中则失败 | 存储引擎限制 |
| `contains` | 字段包含某值则失败 | 禁止SELECT * |
| `gt/lt/gte/lte` | 数值比较 | 操作数量限制 |
| `percent_lt` | 百分比小于则失败 | 注释覆盖率检查 |
| `ratio_gt` | 比率大于则失败 | 删除字段数量检查 |
| `array_len_gt` | 数组长度大于则失败 | JOIN表数量限制 |

### 配置示例

```json
// 1. 主键必需
{"check_field": "table_info.primary_key", "check_type": "empty"}

// 2. 字段注释覆盖率100%
{"check_field": "field_info.comment_summary", "check_type": "percent_lt", "check_value": 100}

// 3. 禁止自定义字符集
{"check_field": "table_info.charset", "check_type": "not_empty"}

// 4. 存储引擎白名单
{"check_field": "table_info.engine", "check_type": "not_in", "check_value": ["InnoDB", ""]}

// 5. UPDATE必须有WHERE
{"check_field": "has_where", "check_type": "equals", "check_value": false}
```

## 默认启用的规则

### 安全类（4条）
1. 禁止删除字段
2. 禁止TRUNCATE
3. UPDATE必须有WHERE
4. DELETE必须有WHERE

### 规范类（6条）
5. 主键必需
6. 表注释必需
7. 字段注释必需（100%）
8. 禁止自定义存储引擎
9. 禁止自定义字符集
10. 禁止自定义排序规则

## 使用方法

### 检查SQL规则

```bash
POST /api/sql/rules/check
Content-Type: application/json

{
  "sql": "CREATE TABLE test (id INT)"
}
```

### 响应格式

```json
{
  "code": 200,
  "message": "SQL规则检查完成",
  "data": {
    "sql_results": [
      {
        "sql": "CREATE TABLE test (id INT)",
        "sql_type": "CREATE",
        "sql_category": "DDL",
        "table_names": ["test"],
        "rule_results": [
          {
            "rule_id": 1,
            "rule_name": "主键必需",
            "rule_type": "field_check",
            "rule_category": "standard",
            "passed": false,
            "error_message": "表必须定义主键"
          }
        ],
        "has_violation": true,
        "has_blocker": true
      }
    ],
    "has_violation": true,
    "has_blocker": true,
    "summary": {
      "total_sql": 1,
      "passed_sql": 0,
      "violation_sql": 1,
      "blocker_sql": 1
    }
  }
}
```

## 扩展规则

新增规则只需在数据库中插入记录：

```sql
INSERT INTO `cmdb`.`sql_rule` 
(`rule_name`, `rule_category`, `rule_type`, `sql_tag`, `rule_description`, `is_enabled`, `rule_content`, `error_message`) 
VALUES
('自定义规则', 'standard', 'field_check', 'CREATE', '规则描述', 1, 
'{"check_field": "table_info.xxx", "check_type": "empty"}', 
'错误提示信息');
```

无需修改代码，规则引擎会自动加载并执行。
