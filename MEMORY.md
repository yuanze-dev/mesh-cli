# Mesh 项目 Memory

## 🤖 AI指令区（AI处理本项目时必读）

**当你读到这个文档时，请按以下顺序执行**：

1. **读取项目宪法**（必须）⭐：
   - 路径：`CONSTITUTION.md`
   - 内容：核心原则、开发规范、产品约束、代码质量标准

2. **理解核心约束**：
   - 无用户注册、无服务器、利用现有同步空间
   - 极简优先、单文件可执行
   - 必须兼容现有 Markdown Memory 系统

3. **阅读协作流程**（必须）⭐：
   - 路径：`docs/04-多AI协作流程.md`
   - 内容：多 AI 协作的工作流程

4. **阅读项目历史**（当前文档）：
   - 了解之前做了什么
   - 理解为什么这样做
   - 避免重复踩坑

5. **阅读技术文档**（根据任务）：
   - 路径：`docs/README.md`
   - 根据任务选择对应模块文档

**执行顺序**：读宪法 → 读协作流程 → 读记忆 → 读技术文档 → 开始工作

---

## 📅 2026-03-13 (v0.1.0)

### 项目初始化 + Git 仓库创建

**核心变更**: 创建 Mesh 项目，实现多 AI 协作的极简记忆同步工具

**原因**:
- 当前基于 Markdown 的 Memory 系统难以多 AI 协作
- 需要 CLI 工具实现结构化存储
- 需要跨电脑、跨 AI 工具的记忆同步
- 想要人 + 多 Agent 协作，实现上下文信息共享

**实施方案**:
- ✅ Go 语言 + SQLite
- ✅ 利用现有同步空间
- ✅ 兼容现有 Markdown Memory 系统
- ✅ 定义多 AI 协作协议

**修改文件**:
- `MEMORY.md` - 创建
- `CONSTITUTION.md` - 创建
- `key.md` - 创建
- `TODO.md` - 人类看的项目进度
- `tasks.md` - AI 看的任务清单
- `docs/` - 完整文档体系

### 项目文档修复

**问题**: MEMORY.md 中路径错误（`sync-space` 应为 `同步空间`）

**修复内容**:
- 修复 MEMORY.md 中所有路径错误
- 修复 CONSTITUTION.md 中的路径错误
- 创建 TODO.md（人类看）和重构 tasks.md（AI 看）
- 更新团队角色定义：人类（晓力）、开发 Agent（多个）、Review Agent（1个）

### Git 仓库创建

**操作内容**:
- 初始化 Git 仓库
- 提交初始代码（12 个文件，2599 行）
- 创建 GitHub 仓库：yuanze-dev/mesh-cli
- 推送代码到远程仓库

**仓库地址**: https://github.com/yuanze-dev/mesh-cli

### 敏感信息检查

**检查结果**: ✅ 无敏感信息泄露

- `key.md` 已被 `.gitignore` 排除，未提交
- 代码中无 token/password/secret 等敏感信息
- git config 中的远程 URL 已修复（移除 token）

### TASK-001 初始化 Go 项目

**核心变更**:
- 新增 `go.mod`，模块名为 `mesh-cli`
- 新增 `cmd/mesh/main.go`，提供 CLI 入口
- 支持 `--help` 与 `--version` 基础参数
- 更新 `TODO.md`，标记 TASK-001 为 ✅ 完成

**修改文件**:
- `go.mod`
- `cmd/mesh/main.go`
- `TODO.md`

**测试结果**:
- 已安装 Go：`go1.26.1 darwin/amd64`
- `go build ./...` 成功
- `go build -o mesh cmd/mesh/main.go` 成功
- `./mesh --help` 输出正常
- `./mesh --version` 输出 `mesh v0.1.0`

---

## 🔧 技术栈总结

**语言**: Go 1.21+
**存储**: SQLite
**构建**: 单文件可执行
**平台**: macOS / Windows / Linux

---

## 🎯 核心设计原则

1. **极简优先**: 核心功能 3-4 个命令，5 分钟上手
2. **零服务器**: 利用现有同步空间，无额外基础设施
3. **零注册**: 无用户体系，设备通过唯一标识区分
4. **兼容现有**: 完美兼容 Markdown Memory 系统
5. **可扩展**: Agent 通过简单 JSON 协议接入

---

## 👥 团队角色

| 角色 | 人数 | 职责 | 工具 |
|-----|------|------|------|
| **人类（晓力）** | 1 | 定义任务、跟进进展、最终验收 | 查看 TODO.md |
| **开发 Agent** | 多个 | 执行具体开发任务 | Claude、Cursor 等 |
| **Review Agent** | 1 | 代码审查、技术把关 | 指定的专业 AI |

---

## 📝 待解决问题

1. **冲突解决**: 多电脑同时编辑同一数据时的冲突处理策略
2. **性能**: 大量数据时的查询性能优化
3. **安全**: 敏感信息的加密存储方案

---

## 🚀 快速开始

### 首次使用

```bash
# 1. 进入项目目录
cd /Users/xiaolin/Downloads/同步空间/mesh

# 2. 初始化
go mod init mesh-cli

# 3. 编译
go build -o mesh cmd/mesh/main.go

# 4. 初始化同步空间
./mesh init --sync-space /Users/xiaolin/Downloads/同步空间/.mesh

# 5. 导出现有 memory（可选）
./mesh import /Users/xiaolin/Downloads/同步空间/Claude code/memory/work/memory.md
```

### 基本命令

```bash
# 采集数据
./mesh collect --source "claude" --content "决策：使用Go开发Mesh" --tag "技术,决策"

# 查询数据
./mesh query "定价策略"

# 同步
./mesh sync

# 导出
./mesh export > memory.md
```

---

**最后更新**: 2026-03-13
**更新人**: Claude Code + 晓力
**当前版本**: v0.1.0
**状态**: 初始化阶段
