// Package agentcmd 实现 mesh agent 命令。
package agentcmd

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"mesh-cli/internal/collect"
	"mesh-cli/internal/store"
	"mesh-cli/pkg/types"
)

const defaultListLimit = 20

// Run 执行 mesh agent 子命令。
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	switch args[0] {
	case "register":
		return runRegister(args[1:], stdout, stderr)
	case "list":
		return runList(args[1:], stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "agent: 不支持的子命令: %s\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runRegister(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("agent register", flag.ContinueOnError)
	fs.SetOutput(stderr)
	id := fs.String("id", "", "Agent ID（必填）")
	name := fs.String("name", "", "Agent 名称（必填）")
	typ := fs.String("type", "", "Agent 类型（可选）")
	device := fs.String("device", "", "设备标识（可选）")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*id) == "" || strings.TrimSpace(*name) == "" {
		_, _ = fmt.Fprintln(stderr, "agent register: 缺少必填参数 --id 或 --name")
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "agent register: 不支持的位置参数: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	dbPath, err := collect.ResolveDBPathForReuse()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "agent register: 解析数据库路径失败: %v\n", err)
		return 1
	}
	st, err := store.NewStore(dbPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "agent register: 初始化存储失败: %v\n", err)
		return 1
	}
	defer func() { _ = st.Close() }()

	a := &types.Agent{ID: strings.TrimSpace(*id), Name: strings.TrimSpace(*name), Type: strings.TrimSpace(*typ), Device: strings.TrimSpace(*device), LastSeen: time.Now().Unix()}
	if err := st.UpsertAgent(a); err != nil {
		_, _ = fmt.Fprintf(stderr, "agent register: 写入失败: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "✅ agent 已注册: %s (%s)\n", a.Name, a.ID)
	return 0
}

func runList(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("agent list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	limit := fs.Int("limit", defaultListLimit, "限制返回数量（默认 20）")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(stderr, "agent list: 不支持的位置参数: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	dbPath, err := collect.ResolveDBPathForReuse()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "agent list: 解析数据库路径失败: %v\n", err)
		return 1
	}
	st, err := store.NewStore(dbPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "agent list: 初始化存储失败: %v\n", err)
		return 1
	}
	defer func() { _ = st.Close() }()

	agents, err := st.ListAgents(*limit)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "agent list: 查询失败: %v\n", err)
		return 1
	}
	if len(agents) == 0 {
		_, _ = fmt.Fprintln(stdout, "暂无 Agent")
		return 0
	}

	_, _ = fmt.Fprintln(stdout, "ID        名称     类型     设备     最后活跃")
	_, _ = fmt.Fprintln(stdout, "------------------------------------------------")
	for _, a := range agents {
		_, _ = fmt.Fprintf(stdout, "%s  %s  %s  %s  %s\n", truncate(a.ID, 8), emptyAsDash(a.Name), emptyAsDash(a.Type), emptyAsDash(a.Device), time.Unix(a.LastSeen, 0).Format("2006-01-02 15:04"))
	}
	_, _ = fmt.Fprintf(stdout, "\n共 %d 个 Agent\n", len(agents))
	return 0
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func emptyAsDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  mesh agent register --id <id> --name <name> [--type <type>] [--device <device>]")
	_, _ = fmt.Fprintln(w, "  mesh agent list [--limit <n>]")
}
