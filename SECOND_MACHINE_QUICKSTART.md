# Mesh 第二台电脑快速接入

在第二台电脑，把 `mesh` 项目目录同步下来后执行：

```bash
cd "/你的/mesh目录"
./setup-second-machine.sh
```

如果你的同步盘路径不是默认 `~/Downloads/同步空间/.mesh-sync`，传入路径参数：

```bash
./setup-second-machine.sh "/你的同步盘路径/.mesh-sync"
```

## 无 Go 环境怎么办

脚本会优先使用本目录下的 `./mesh` 可执行文件。

- 如果 `./mesh` 存在：可直接用，不需要 Go。
- 如果 `./mesh` 不存在且没有 Go：脚本会提示并退出。

你可以在第一台电脑先编译后拷贝二进制：

```bash
cd "/Users/yuanze/Downloads/同步空间/mesh"
go build -o mesh ./cmd/mesh
```

然后把该 `mesh` 文件带到第二台电脑同目录即可。
