# Mesh CLI

极简多 Agent 记忆采集、查询与同步工具。

## 核心能力

- `collect/query/list`：本地 SQLite 记忆库采集与检索
- `import/export`：Markdown Memory 兼容迁移
- `init/push/pull/sync`：多设备文件级同步
- `agent/inject`：多 AI 协作上下文注入

## 快速开始

```bash
go build -o mesh ./cmd/mesh
./mesh --help
./mesh collect --source "claude" --content "决策：使用Go开发" --tag "技术,决策"
./mesh query "决策" --limit 5
```

完整文档见 [docs/README.md](docs/README.md)。

## 文档索引

- [快速入门](docs/01-快速入门.md)
- [命令说明](docs/02-命令说明.md)
- [Agent 协议](docs/03-Agent协议.md)
- [多 AI 协作流程](docs/04-多AI协作流程.md)
- [发布确认单](docs/05-发布确认单-v0.1.x.md)

## 版本与变更

- 当前版本：`v0.1.0`（候选）
- 变更记录：见 [CHANGELOG.md](CHANGELOG.md)

