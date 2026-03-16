#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEFAULT_SYNC="$HOME/Downloads/同步空间/.mesh-sync"
SYNC_PATH="${1:-$DEFAULT_SYNC}"

cd "$PROJECT_DIR"

echo "[1/5] 检查 mesh 可执行文件..."
if [[ ! -x "$PROJECT_DIR/mesh" ]]; then
  if command -v go >/dev/null 2>&1; then
    echo "未检测到可执行文件，使用 Go 编译..."
    go build -o mesh ./cmd/mesh
  else
    cat <<'MSG'
未检测到 Go 环境，也没有可执行文件 ./mesh。
请任选其一：
1) 在有 Go 的机器编译后，把 mesh 二进制拷贝到本目录
2) 安装 Go 后重新执行本脚本
MSG
    exit 1
  fi
fi

echo "[2/5] 检查同步目录: $SYNC_PATH"
mkdir -p "$SYNC_PATH"

echo "[3/5] 初始化配置"
"$PROJECT_DIR/mesh" init --sync-space "$SYNC_PATH"

echo "[4/5] 拉取远端数据"
"$PROJECT_DIR/mesh" pull || true

echo "[5/5] 验证状态"
"$PROJECT_DIR/mesh" --version
"$PROJECT_DIR/mesh" list --limit 5 || true

cat <<MSG

完成。
后续建议流程：
- 开始工作前：./mesh pull
- 记录记忆：  ./mesh collect --source "电脑名" --content "..." --tag "..."
- 结束工作前：./mesh push

如同步目录不是默认值，请这样运行：
./setup-second-machine.sh "/你的同步盘路径/.mesh-sync"
MSG
