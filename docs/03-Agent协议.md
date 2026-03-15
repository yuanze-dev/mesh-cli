# Mesh Agent 协议（TASK-011）

> 目标：定义 Mesh 中多 Agent 协作的最小可用协议，覆盖注册、采集、查询、同步四类消息。

---

## 1. 设计目标

- **极简**：只保留核心字段，默认 JSON。
- **可追踪**：每条消息必须可关联（`request_id`）和可审计（`timestamp`）。
- **兼容 CLI**：协议字段与现有命令语义对齐（`agent/collect/query/sync`）。
- **可扩展**：允许 `metadata` / `extensions` 扩展，不破坏主协议。

---

## 2. 通用信封（Envelope）

### 2.1 请求格式

```json
{
  "version": "1.0",
  "request_id": "req_20260315_0001",
  "message_type": "register|collect|query|sync",
  "agent_id": "claude_main_xiaoli",
  "timestamp": 1773537600,
  "payload": {},
  "metadata": {
    "trace_id": "optional-trace-id",
    "client": "mesh-cli"
  }
}
```

### 2.2 响应格式

```json
{
  "version": "1.0",
  "request_id": "req_20260315_0001",
  "status": "success|error",
  "timestamp": 1773537601,
  "data": {},
  "error": {
    "code": "",
    "message": "",
    "details": {}
  }
}
```

### 2.3 字段定义

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `version` | string | 是 | 协议版本，当前固定 `1.0` |
| `request_id` | string | 是 | 请求唯一 ID，用于链路追踪与幂等 |
| `message_type` | string | 是 | `register`/`collect`/`query`/`sync` |
| `agent_id` | string | 是 | 发送方 Agent ID |
| `timestamp` | int64 | 是 | Unix 秒时间戳 |
| `payload` | object | 是 | 与消息类型相关的业务内容 |
| `metadata` | object | 否 | 可扩展元数据（trace、client、env 等） |

---

## 3. 消息类型定义

## 3.1 `register`（Agent 注册）

### 请求 `payload`

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | string | 是 | Agent 展示名 |
| `type` | string | 否 | Agent 类型（claude/cursor/chatgpt/agent...） |
| `device` | string | 否 | 设备标识 |
| `capabilities` | string[] | 否 | 能力声明，如 `collect`,`query`,`sync` |
| `extensions` | object | 否 | 自定义扩展字段 |

### 请求示例

```json
{
  "version": "1.0",
  "request_id": "req_register_001",
  "message_type": "register",
  "agent_id": "claude_main_xiaoli",
  "timestamp": 1773537600,
  "payload": {
    "name": "Claude Main",
    "type": "claude",
    "device": "mini2",
    "capabilities": ["collect", "query", "sync"]
  }
}
```

---

## 3.2 `collect`（采集认知）

### 请求 `payload`

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `source` | string | 是 | 记录来源 |
| `content` | string | 是 | 认知正文 |
| `summary` | string | 否 | 摘要 |
| `tags` | string | 否 | 标签，逗号分隔 |
| `is_shared` | int | 否 | 0/1，是否可共享 |

### 请求示例

```json
{
  "version": "1.0",
  "request_id": "req_collect_001",
  "message_type": "collect",
  "agent_id": "claude_main_xiaoli",
  "timestamp": 1773537600,
  "payload": {
    "source": "claude_main_xiaoli",
    "content": "完成 TASK-011 协议定稿并更新文档",
    "summary": "TASK-011 完成",
    "tags": "mesh,协议,review",
    "is_shared": 1
  }
}
```

---

## 3.3 `query`（查询认知）

### 请求 `payload`

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `query` | string | 否 | 关键词 |
| `source` | string | 否 | 来源过滤 |
| `tags` | string | 否 | 标签过滤 |
| `limit` | int | 否 | 返回上限（建议 1~100） |

> 约束：`query/source/tags` 至少提供一个。

### 请求示例

```json
{
  "version": "1.0",
  "request_id": "req_query_001",
  "message_type": "query",
  "agent_id": "cursor_worker_1",
  "timestamp": 1773537600,
  "payload": {
    "query": "同步冲突",
    "source": "",
    "tags": "sync",
    "limit": 5
  }
}
```

---

## 3.4 `sync`（同步状态/动作）

### 请求 `payload`

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `action` | string | 是 | `push` / `pull` / `sync` |
| `sync_space` | string | 否 | 同步空间路径（可选） |
| `db_path` | string | 否 | 数据库路径（可选） |
| `dry_run` | bool | 否 | 是否仅预演 |

### 请求示例

```json
{
  "version": "1.0",
  "request_id": "req_sync_001",
  "message_type": "sync",
  "agent_id": "claude_main_xiaoli",
  "timestamp": 1773537600,
  "payload": {
    "action": "sync",
    "dry_run": false
  }
}
```

---

## 4. Agent ID 规范

格式：`{type}_{name}_{device}`

示例：
- `claude_main_mini2`
- `cursor_worker1_macbook`
- `agent_backend_server`

约束：
- 小写字母、数字、下划线
- 长度建议 3~64
- 全局唯一（至少在同一同步空间内唯一）

---

## 5. 错误码规范

| 错误码 | 含义 |
|---|---|
| `INVALID_MESSAGE_TYPE` | `message_type` 非法 |
| `MISSING_REQUIRED_FIELD` | 缺少必填字段 |
| `INVALID_AGENT_ID` | Agent ID 不合法或未注册 |
| `INVALID_PAYLOAD` | payload 格式错误 |
| `DATABASE_ERROR` | 存储层异常 |
| `SYNC_ERROR` | 同步失败 |
| `INTERNAL_ERROR` | 未知内部错误 |

错误响应示例：

```json
{
  "version": "1.0",
  "request_id": "req_sync_001",
  "status": "error",
  "timestamp": 1773537601,
  "error": {
    "code": "SYNC_ERROR",
    "message": "sync lock busy",
    "details": {
      "lock": ".mesh.sync.lock"
    }
  }
}
```

---

## 6. 与当前 CLI 的映射

| 协议消息 | Mesh CLI 命令 |
|---|---|
| `register` | `mesh agent register` |
| `collect` | `mesh collect` |
| `query` | `mesh query` |
| `sync(action=push/pull/sync)` | `mesh push` / `mesh pull` / `mesh sync` |

---

## 7. TASK-011 验收清单（Review Agent）

- [x] 定义统一 Envelope（请求/响应）
- [x] 定义四类消息：`register`/`collect`/`query`/`sync`
- [x] 定义字段级约束与错误码
- [x] 提供可直接复用的 JSON 示例
- [x] 给出与现有 CLI 的映射关系

**结论**：TASK-011 协议文档已完成，可进入验收。

---

**最后更新**: 2026-03-15
**版本**: 1.1
