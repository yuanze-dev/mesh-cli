# Mesh AI 任务清单

> AI 看的任务详情 | 每完成一个任务，更新 TODO.md 中的状态

---

## 🤖 给 AI 的说明

**你是 Mesh 项目的开发 Agent 或 Review Agent。**

### 执行流程

1. **读取宪法**（必须）：
   ```bash
   cat CONSTITUTION.md
   ```

2. **阅读协作流程**（必须）：
   ```bash
   cat docs/04-多AI协作流程.md
   ```

3. **找到你的任务**：
   - 查看下方任务列表
   - 根据任务 ID 找到你的任务

4. **执行任务**：
   - 按照任务要求开发
   - 遵循宪法中的规范

5. **更新状态**：
   - 开发 Agent 完成 → 更新 TODO.md
   - Review Agent 审查 → 更新 TODO.md

### 角色说明

| 角色 | 任务范围 |
|-----|---------|
| **开发 Agent** | TASK-001 到 TASK-016（除 011） |
| **Review Agent** | TASK-011（协议定义）、所有代码审查 |

---

## 第一阶段：核心存储层

### TASK-001 初始化 Go 项目

**角色**: 开发 Agent
**优先级**: P0
**状态**: ⏳ 待开始

**任务描述**:
初始化 Go 项目，设置基础结构和依赖。

**验收标准**:
- [ ] `go mod init mesh-cli` 完成
- [ ] 目录结构符合规范
- [ ] `go build` 能成功编译
- [ ] 基础的 `--help` 命令可用

**目录结构**（已创建）:
```
/Users/xiaolin/Downloads/同步空间/mesh/
├── cmd/mesh/         # CLI 入口
├── internal/          # 内部包
├── pkg/               # 公共包
└── docs/              # 文档
```

**需要创建的文件**:
1. `go.mod` - 初始化模块
2. `cmd/mesh/main.go` - CLI 入口
3. `.gitignore` - 忽略 build 文件

**完成更新**:
更新 `TODO.md` 中 TASK-001 状态为 ✅ 完成

---

### TASK-002 实现 SQLite 存储层

**角色**: 开发 Agent
**优先级**: P0
**状态**: ⏳ 待开始
**依赖**: TASK-001

**任务描述**:
实现 SQLite 存储层，包括数据库初始化、CRUD 操作。

**验收标准**:
- [ ] 数据库文件自动创建
- [ ] 支持插入记录（Insert）
- [ ] 支持查询记录（Query）
- [ ] 支持列出记录（List）
- [ ] 有单元测试

**文件位置**:
`internal/store/sqlite.go`

**数据库结构**:
```sql
CREATE TABLE insights (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    content TEXT NOT NULL,
    summary TEXT,
    tags TEXT,
    agent_id TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER,
    is_shared INTEGER DEFAULT 0
);
```

**数据结构**（已定义）:
`pkg/types/types.go` 中的 `Insight` 结构体

**需要实现的方法**:
1. `NewStore(dbPath string) (*Store, error)` - 初始化存储
2. `Insert(insight *Insight) error` - 插入记录
3. `Query(query string, limit int) ([]*Insight, error)` - 查询记录
4. `List(limit int) ([]*Insight, error)` - 列出记录
5. `Close() error` - 关闭连接

**完成更新**:
更新 `TODO.md` 中 TASK-002 状态为 ✅ 完成

---

### TASK-003 实现 mesh collect 命令

**角色**: 开发 Agent
**优先级**: P0
**状态**: ⏳ 待开始
**依赖**: TASK-002
**需要 Review**: 是

**任务描述**:
实现 `mesh collect` 命令，用于采集数据。

**验收标准**:
- [ ] 支持 `--source` 参数
- [ ] 支持 `--content` 参数
- [ ] 支持 `--tag` 参数
- [ ] 数据成功写入 SQLite
- [ ] 有友好的输出提示

**文件位置**:
`internal/collect/collect.go`

**命令示例**:
```bash
./mesh collect --source "claude" --content "决策：使用Go开发" --tag "技术,决策"
```

**实现流程**:
1. 解析命令行参数
2. 验证必填参数
3. 生成 UUID 作为 ID
4. 调用 `store.Insert()` 保存
5. 输出友好提示

**输出示例**:
```
✅ 已采集 1 条认知
   来源: claude
   标签: 技术,决策
```

**完成后**:
1. 更新 `TODO.md` 中 TASK-003 状态为 ⏳ 待审查
2. 提交给 Review Agent 审查

---

### TASK-004 实现 mesh query 命令

**角色**: 开发 Agent
**优先级**: P0
**状态**: ⏳ 待开始
**依赖**: TASK-002
**需要 Review**: 是

**任务描述**:
实现 `mesh query` 命令，用于查询数据。

**验收标准**:
- [ ] 支持关键词查询
- [ ] 支持 `--source` 过滤
- [ ] 支持 `--tag` 过滤
- [ ] 支持 `--limit` 限制结果数
- [ ] 输出格式化友好

**命令示例**:
```bash
./mesh query "定价策略"
./mesh query --source "claude" --tag "产品"
./mesh query --limit 5 "最近"
```

**完成后**:
1. 更新 `TODO.md` 中 TASK-004 状态为 ⏳ 待审查
2. 提交给 Review Agent 审查

---

### TASK-005 实现 mesh list 命令

**角色**: 开发 Agent
**优先级**: P0
**状态**: ⏳ 待开始
**依赖**: TASK-002
**需要 Review**: 是

**任务描述**:
实现 `mesh list` 命令，列出所有记录。

**验收标准**:
- [ ] 列出最近 N 条记录
- [ ] 支持 `--source` 过滤
- [ ] 支持 `--limit` 限制数量
- [ ] 表格格式输出

**命令示例**:
```bash
./mesh list
./mesh list --source "claude" --limit 10
```

**完成后**:
1. 更新 `TODO.md` 中 TASK-005 状态为 ⏳ 待审查
2. 提交给 Review Agent 审查

---

## 第二阶段：Markdown 兼容

### TASK-006 实现 Markdown 导入

**角色**: 开发 Agent
**优先级**: P0
**状态**: ⏳ 待开始
**需要 Review**: 是

**任务描述**:
实现从 Markdown 文件导入数据到 Mesh。

**验收标准**:
- [ ] 能解析 Markdown 格式
- [ ] 提取内容、标签、时间
- [ ] 导入到 SQLite
- [ ] 跳过重复内容

**Markdown 格式示例**:
```markdown
## 2026-03-10

### 决策：使用 Go 开发 Mesh
标签: 技术,决策
内容: xxx
```

**命令示例**:
```bash
./mesh import /path/to/memory.md
```

**完成后**:
更新 `TODO.md` 中 TASK-006 状态为 ⏳ 待审查

---

### TASK-007 实现 Markdown 导出

**角色**: 开发 Agent
**优先级**: P0
**状态**: ⏳ 待开始
**需要 Review**: 是

**任务描述**:
实现从 Mesh 导出数据为 Markdown 格式。

**验收标准**:
- [ ] 导出为标准 Markdown
- [ ] 保留原有格式
- [ ] 支持输出到文件或 stdout

**命令示例**:
```bash
./mesh export > memory.md
./mesh export --output memory.md
```

**完成后**:
更新 `TODO.md` 中 TASK-007 状态为 ⏳ 待审查

---

## 第三阶段：同步功能

### TASK-008 实现 mesh init 命令

**角色**: 开发 Agent
**优先级**: P1
**状态**: ⏳ 待开始

**任务描述**:
实现初始化配置，设置同步空间路径。

**验收标准**:
- [ ] 创建配置文件
- [ ] 设置同步空间路径
- [ ] 创建必要的目录
- [ ] 验证配置正确

**配置文件位置**: `~/.mesh/config.json`

**完成后**:
更新 `TODO.md` 中 TASK-008 状态为 ⏳ 待审查

---

### TASK-009 实现 mesh push/pull 命令

**角色**: 开发 Agent
**优先级**: P1
**状态**: ⏳ 待开始

**任务描述**:
实现推送到同步空间和从同步空间拉取。

**验收标准**:
- [ ] push 复制本地 DB 到同步空间
- [ ] pull 复制同步空间 DB 到本地
- [ ] 有文件锁，防止并发冲突

**完成后**:
更新 `TODO.md` 中 TASK-009 状态为 ⏳ 待审查

---

### TASK-010 实现 mesh sync 命令

**角色**: 开发 Agent
**优先级**: P1
**状态**: ⏳ 待开始

**任务描述**:
实现双向同步（push + pull）。

**验收标准**:
- [ ] 先 pull 再 push
- [ ] 简单的冲突处理
- [ ] 显示同步结果

**完成后**:
更新 `TODO.md` 中 TASK-010 状态为 ⏳ 待审查

---

## 第四阶段：Agent 协作

### TASK-011 定义 Agent 协议

**角色**: Review Agent
**优先级**: P2
**状态**: ⏳ 待开始

**任务描述**:
定义 Agent 协议格式和消息类型。

**验收标准**:
- [ ] 定义协议格式
- [ ] 定义消息类型（register/collect/query/sync）
- [ ] 编写协议文档

**说明**:
- 协议文档已创建：`docs/03-Agent协议.md`
- Review Agent 需要审查和完善协议

**完成后**:
更新 `TODO.md` 中 TASK-011 状态为 ✅ 完成

---

### TASK-012 实现 mesh agent 命令

**角色**: 开发 Agent
**优先级**: P2
**状态**: ⏳ 待开始

**任务描述**:
实现 Agent 注册和管理命令。

**验收标准**:
- [ ] `mesh agent register` - 注册 Agent
- [ ] `mesh agent list` - 列出 Agent
- [ ] 存储到 SQLite

**完成后**:
更新 `TODO.md` 中 TASK-012 状态为 ⏳ 待审查

---

### TASK-013 实现 mesh inject 命令

**角色**: 开发 Agent
**优先级**: P2
**状态**: ⏳ 待开始

**任务描述**:
实现上下文注入命令，给其他 AI 提供上下文。

**验收标准**:
- [ ] 查询相关认知
- [ ] 格式化为可读文本
- [ ] 支持 `--max-tokens` 限制
- [ ] 支持 `--format` 输出格式

**完成后**:
更新 `TODO.md` 中 TASK-013 状态为 ⏳ 待审查

---

## 第五阶段：文档与打包

### TASK-014 编写快速入门文档

**角色**: 开发 Agent
**优先级**: P1
**状态**: ⏳ 待开始

**任务描述**:
编写快速入门文档，让用户 5 分钟上手。

**验收标准**:
- [ ] 安装步骤清晰
- [ ] 基本命令示例
- [ ] 常见问题解答
- [ ] 文件位置：`docs/01-快速入门.md`

**说明**:
文档已创建，开发 Agent 需要根据实际实现更新内容

**完成后**:
更新 `TODO.md` 中 TASK-014 状态为 ✅ 完成

---

### TASK-015 编写命令说明文档

**角色**: 开发 Agent
**优先级**: P1
**状态**: ⏳ 待开始

**任务描述**:
编写完整的命令说明文档。

**验收标准**:
- [ ] 所有命令都有说明
- [ ] 参数说明完整
- [ ] 示例代码清晰
- [ ] 文件位置：`docs/02-命令说明.md`

**说明**:
文档已创建，开发 Agent 需要根据实际实现更新内容

**完成后**:
更新 `TODO.md` 中 TASK-015 状态为 ✅ 完成

---

### TASK-016 打包为可执行文件

**角色**: Review Agent
**优先级**: P1
**状态**: ⏳ 待开始

**任务描述**:
编译打包为单文件可执行。

**验收标准**:
- [ ] macOS Intel 版
- [ ] macOS Apple Silicon 版
- [ ] Windows x64 版

**完成后**:
更新 `TODO.md` 中 TASK-016 状态为 ✅ 完成

---

## 🔍 状态说明

| 状态 | 说明 |
|-----|------|
| ⏳ 待开始 | 任务还未开始 |
| ⏳ 待审查 | 开发完成，等待 Review Agent 审查 |
| ✅ 完成 | 任务已通过审查 |
| ❌ 阻塞 | 任务被阻塞，等待依赖完成 |

---

## 📌 给 Review Agent 的审查清单

每个需要审查的任务，Review Agent 必须检查：

- [ ] 代码质量（规范、注释、复杂度）
- [ ] 功能完整性（按需求实现）
- [ ] 测试覆盖（核心功能有测试）
- [ ] 文档更新（相关文档已更新）

审查通过后：
1. 更新 `TODO.md` 中任务状态为 ✅ 完成
2. 通知晓力验收

---

**最后更新**: 2026-03-13
