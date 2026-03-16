# Changelog

All notable changes to this project are documented in this file.

## [v0.1.0] - 2026-03-15

### Added

- 完整 CLI 子命令：`collect/query/list/import/export/init/push/pull/sync/agent/inject`
- SQLite 存储层与核心单元测试
- Markdown Memory 导入导出与去重逻辑
- Agent 注册与上下文注入功能
- 多 AI 协作文档与发布确认单

### Fixed

- 主入口补齐 `query` 子命令，`mesh --help` 与实际命令一致
- 文档路径与导航链接更新为当前仓库结构

### Verified

- 2026-03-15 全量单元测试：`go test ./...` 通过
- 2026-03-15 端到端命令链路验证通过（collect/query/list/import/export/init/push/pull/sync/agent/inject）

