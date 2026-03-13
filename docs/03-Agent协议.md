# Mesh Agent 协议

> Agent 接入 Mesh 的标准协议

---

## 一、协议概述

Mesh Agent 协议是一个极简的 JSON 协议，用于 AI Agent 与 Mesh 之间的通信。

**核心原则**：
- 极简：只有 4 种消息类型
- 无状态：每次请求独立
- 可扩展：metadata 字段支持任意扩展

---

## 二、协议格式

### 2.1 基础格式

```json
{
  "version": "1.0",
  "message_type": "register|collect|query|heartbeat",
  "agent_id": "agent_identifier",
  "timestamp": 1678900000,
  "payload": {
    // 消息类型特定内容
  }
}
```

### 2.2 字段说明

| 字段 | 类型 | 必填 | 说明 |
|-----|------|------|------|
| `version` | string | 是 | 协议版本，当前 "1.0" |
| `message_type` | string | 是 | 消息类型 |
| `agent_id` | string | 是 | Agent 唯一标识 |
| `timestamp` | int64 | 是 | Unix 时间戳（秒） |
| `payload` | object | 是 | 消息内容 |

---

## 三、消息类型

### 3.1 register - Agent 注册

**用途**: Agent 首次启动时注册自身信息

**请求示例**:
```json
{
  "version": "1.0",
  "message_type": "register",
  "agent_id": "cursor_main_shawn",
  "timestamp": 1678900000,
  "payload": {
    "name": "Cursor Main",
    "type": "cursor",
    "device": "shawn",
    "metadata": {
      "version": "1.0",
      "capabilities": ["collect", "query"]
    }
  }
}
```

**响应示例**:
```json
{
  "status": "success",
  "message": "Agent registered successfully"
}
```

---

### 3.2 collect - 采集认知

**用途**: Agent 完成任务后，上报认知

**请求示例**:
```json
{
  "version": "1.0",
  "message_type": "collect",
  "agent_id": "cursor_main_shawn",
  "timestamp": 1678900000,
  "payload": {
    "source": "cursor_main_shawn",
    "content": "完成登录功能开发，使用 JWT 认证",
    "summary": "登录功能已完成",
    "tags": "开发,完成,登录",
    "is_shared": 1
  }
}
```

**响应示例**:
```json
{
  "status": "success",
  "insight_id": "uuid-xxx",
  "message": "Insight collected successfully"
}
```

---

### 3.3 query - 查询认知

**用途**: Agent 开始任务前，查询相关认知

**请求示例**:
```json
{
  "version": "1.0",
  "message_type": "query",
  "agent_id": "cursor_main_shawn",
  "timestamp": 1678900000,
  "payload": {
    "query": "登录功能",
    "source": "",
    "tags": "开发",
    "limit": 5
  }
}
```

**响应示例**:
```json
{
  "status": "success",
  "results": [
    {
      "id": "uuid-xxx",
      "source": "claude_main_xiaolin",
      "content": "决策：使用 JWT 认证方式",
      "summary": "JWT 认证决策",
      "tags": "开发,认证",
      "created_at": 1678890000
    },
    {
      "id": "uuid-yyy",
      "source": "chatgpt_main",
      "content": "竞品登录流程分析...",
      "summary": "登录流程分析",
      "tags": "开发,产品",
      "created_at": 1678880000
    }
  ]
}
```

---

### 3.4 heartbeat - 心跳

**用途**: Agent 定期报告存活状态

**请求示例**:
```json
{
  "version": "1.0",
  "message_type": "heartbeat",
  "agent_id": "cursor_main_shawn",
  "timestamp": 1678900000,
  "payload": {}
}
```

**响应示例**:
```json
{
  "status": "success",
  "message": "Heartbeat received",
  "last_seen": 1678900000
}
```

---

## 四、Agent ID 规范

### 4.1 格式

```
{类型}_{名称}_{设备}
```

### 4.2 示例

| Agent ID | 类型 | 名称 | 设备 |
|----------|------|------|------|
| `claude_main_xiaolin` | claude | main | xiaolin |
| `claude_worker_1_xiaolin` | claude | worker_1 | xiaolin |
| `cursor_main_shawn` | cursor | main | shawn |
| `agent_backend_server` | agent | backend | server |

### 4.3 类型规范

| 类型 | 说明 |
|-----|------|
| `claude` | Claude 系列 AI |
| `chatgpt` | ChatGPT 系列 AI |
| `cursor` | Cursor IDE |
| `claude-code` | Claude Code CLI |
| `agent` | 自定义 Agent |

---

## 五、集成示例

### 5.1 Claude Code Skill 集成

```bash
# 开始任务前
CONTEXT=$(mesh query "定价策略" --format json)

# 执行任务...

# 任务完成后
mesh collect \
  --source "claude-code" \
  --content "完成定价功能开发" \
  --tag "开发,完成"
```

### 5.2 Python Agent 集成

```python
import subprocess
import json

def query_context(query):
    result = subprocess.run(
        ["mesh", "query", query, "--format", "json"],
        capture_output=True,
        text=True
    )
    return json.loads(result.stdout)

def collect_insight(content, tags):
    subprocess.run([
        "mesh", "collect",
        "--source", "python_agent",
        "--content", content,
        "--tag", tags
    ])

# 使用
context = query_context("登录功能")
# ... 执行任务 ...
collect_insight("完成登录开发", "开发,完成")
```

### 5.3 JavaScript/Node.js 集成

```javascript
const { execSync } = require('child_process');

function queryContext(query) {
    const result = execSync(`mesh query "${query}" --format json`);
    return JSON.parse(result.toString());
}

function collectInsight(content, tags) {
    execSync(`mesh collect --source "node_agent" --content "${content}" --tag "${tags}"`);
}

// 使用
const context = queryContext('登录功能');
// ... 执行任务 ...
collectInsight('完成登录开发', '开发,完成');
```

---

## 六、最佳实践

### 6.1 心跳频率

建议心跳间隔：**5-10 分钟**

### 6.2 认知采集时机

- **任务完成后**：立即采集
- **重要决策后**：立即采集
- **发现新信息后**：立即采集

### 6.3 上下文查询时机

- **开始任务前**：查询相关上下文
- **遇到问题前**：查询历史解决方案
- **需要参考前**：查询相关认知

### 6.4 标签使用规范

| 场景 | 推荐标签 |
|-----|---------|
| 开发任务 | `开发` |
| 完成任务 | `完成` |
| 产品决策 | `产品,决策` |
| 技术决策 | `技术,决策` |
| Bug 修复 | `bug,修复` |
| 测试相关 | `测试` |

---

## 七、错误处理

### 7.1 错误响应格式

```json
{
  "status": "error",
  "error_code": "INVALID_AGENT_ID",
  "message": "Agent ID not found"
}
```

### 7.2 错误码

| 错误码 | 说明 |
|-------|------|
| `INVALID_AGENT_ID` | Agent ID 不存在或未注册 |
| `MISSING_REQUIRED_FIELD` | 缺少必填字段 |
| `INVALID_MESSAGE_TYPE` | 消息类型无效 |
| `DATABASE_ERROR` | 数据库错误 |
| `SYNC_ERROR` | 同步错误 |

---

**最后更新**: 2026-03-13
**版本**: 1.0
